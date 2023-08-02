/*
Copyright © 2023 XIAOMIN LAI
*/

package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

func gitCloneRevision(repo Repository) (*git.Repository, error) {

	// Create credentials
	creds := &http.BasicAuth{
		Username: repo.Username,
		Password: repo.Password,
	}

	// Init memory storage and fs
	storer := memory.NewStorage()
	fs := memfs.New()

	// Clone repo into memfs
	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:  repo.Url,
		Auth: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("could not git clone: %w", err)
	}

	// Get git default worktree
	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("could not get git worktree: %w", err)
	}

	// If revision is empty, pull master or main branch, otherwise return error
	if repo.Revision == "" {
		err = w.Pull(&git.PullOptions{
			RemoteName:      git.DefaultRemoteName,
			Auth:            creds,
			InsecureSkipTLS: true,
		})
		if err != nil {
			return nil, fmt.Errorf("could not git pull master or main: %w", err)
		}

	} else {
		// Pull the revision
		err = w.Pull(&git.PullOptions{
			RemoteName:      git.DefaultRemoteName,
			ReferenceName:   plumbing.NewBranchReferenceName(repo.Revision),
			Auth:            creds,
			Force:           true,
			InsecureSkipTLS: true, // TODO: implement repo.Insecure
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("could not git pull the given revision: %w", err)
		}

		// Create a new reference to the pulled revision
		commitStr, err := r.ResolveRevision(plumbing.Revision(plumbing.NewRemoteReferenceName(git.DefaultRemoteName, repo.Revision)))
		if err != nil {
			return nil, fmt.Errorf("could not resolve the given revision: %w", err)
		}
		log.Debugf("Resolved revision %s to %s", repo.Revision, commitStr.String())

		branchRef := plumbing.NewReferenceFromStrings(plumbing.NewBranchReferenceName(repo.Revision).String(), commitStr.String())
		err = r.Storer.SetReference(branchRef)
		if err != nil {
			return nil, fmt.Errorf("could not set reference from string: %w", err)
		}
		log.Debugf("Created branch %s at %s", repo.Revision, commitStr.String())

		// Checkout the revision (HEAD is now at the revision)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(repo.Revision),
		})
		if err != nil {
			return nil, fmt.Errorf("could not git checkout the given revision: %w", err)
		}

		ref, _ := r.Head()
		log.Debugf("Now HEAD is at %s", ref.String())

	}

	return r, nil
}

func getFileList(r *git.Repository, path string) ([]string, error) {

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	log.Debugf("output ref: %#v", ref)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("could not get HEAD: %w", err)
	}

	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	log.Debugf("output commit id: %#v", commit.ID().String())
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("could not get commit object: %w", err)
	}

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	log.Tracef("output tree: %#v", tree)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve tree of commit: %w", err)
	}

	files := make([]string, 0)

	// ... get the files iterator
	err = tree.Files().ForEach(func(f *object.File) error {
		if yes, err := filepath.Match(path, f.Name); yes && err == nil {
			log.Debugf("file: %s, hash: %s", f.Name, f.Hash)
			files = append(files, f.Name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not git ls-files: %w", err)
	}

	log.Debugf("getFileList output: %+v\n", files)

	return files, nil
}

func gitDiff(r *git.Repository, oldCommitID, newCommitID string) (*object.Patch, error) {

	// retrieve the commit object from old commit id
	from, err := r.CommitObject(plumbing.NewHash(oldCommitID))
	if err != nil {
		return nil, fmt.Errorf("could not get commit object from old commit id: %w", err)
	}

	// retrieve the commit object from new commit id
	to, err := r.CommitObject(plumbing.NewHash(newCommitID))
	if err != nil {
		return nil, fmt.Errorf("could not get commit object from new commit id: %w", err)
	}

	// get the patch
	patch, err := from.Patch(to)
	if err != nil {
		return nil, fmt.Errorf("could not get patch: %w", err)
	}

	return patch, nil

}

type fileStatus int

const (
	CREATED fileStatus = iota
	DELETED
)

type fileStat struct {
	Name string
	Stat fileStatus
}

func getFilePathAndStatus(filePatch diff.FilePatch) []*fileStat {

	// get the from and to file
	from, to := filePatch.Files()

	output := make([]*fileStat, 0)

	// check if the file is created
	if from == nil {
		return append(output, &fileStat{
			Name: to.Path(),
			Stat: CREATED,
		})
	}

	// check if the file is deleted
	if to == nil {
		return append(output, &fileStat{
			Name: from.Path(),
			Stat: DELETED,
		})
	}

	// check if the file is modified
	if from != nil && to != nil && from.Path() == to.Path() {
		return output
	}

	// check if the file is renamed
	if from != nil && to != nil && from.Path() != to.Path() {
		output = append(output, &fileStat{
			Name: from.Path(),
			Stat: DELETED,
		})
		output = append(output, &fileStat{
			Name: to.Path(),
			Stat: CREATED,
		})
		return output
	}

	return output
}
