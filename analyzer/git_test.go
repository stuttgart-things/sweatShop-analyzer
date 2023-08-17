package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_gitDiff(t *testing.T) {

	repo := &Repository{
		Name: "testing-repo",
		Url:  "https://github.com/go-git/go-git",
		// Url: "https://github.com/aws-samples/eks-gitops-crossplane-argocd",
		// Revision: "master",
		Revision: "wasm",
		// Revision: "main",
	}

	gitRepo, _ := gitCloneRevision(repo)
	currentCommitID, _ := gitRepo.Head()

	// Create a new redis cache
	cache := newAnalyzerCacheForTesting()
	cached, _ := cache.GetMatchingFiles(repo.Url)

	_, err := gitDiff(gitRepo, cached.CommitID, currentCommitID.Hash().String())
	assert.NoError(t, err)

	assert.Fail(t, "Intentionally failing test to see output")
}
