/*
Copyright © 2023 XIAOMIN LAI
*/

package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	memfs "github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
	goredis "github.com/redis/go-redis/v9"

	sthingsBase "github.com/stuttgart-things/sthingsBase"
)

type Repository struct {
	Name                  string
	Url                   string
	Revision              string
	Username              string
	Password              string
	Insecure              bool
	ForceCompleteAnalysis *bool
}

// TechAndPath is a map with technology and a path
type TechAndPath struct {
	// name of the technology
	Technology string
	// path, that matches to the pattern of the technology
	Path string
}

var (
	redisServer = os.Getenv("REDIS_SERVER")
	log         = sthingsBase.StdOutFileLogger(logfilePath, "2006-01-02 15:04:05", 50, 3, 28)
	logfilePath = "/tmp/sweatShop-analyzer.log"
)

// ConnectRepository tests the repository connection and authentication
func (repo *Repository) ConnectRepository() error {
	// Init memory storage and fs
	storer := memory.NewStorage()
	fs := memfs.New()

	// Clone repo into memfs
	_, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: repo.Url,
		Auth: &http.BasicAuth{
			Username: repo.Username,
			Password: repo.Password,
		},
	})
	if err != nil {
		return fmt.Errorf("could not git clone repository %s: %w", repo.Url, err)
	}
	log.Infoln("Repository cloned")

	return nil
}

func (repo *Repository) GetMatchingFiles() {
	log.Println("sweatShop-analyzer started")
	log.Println("GetMatchingFiles for repo:", repo)

	// Clone the repo into memory for later use
	gitRepo, err := gitCloneRevision(repo)

	if err != nil {
		log.Errorf("could not clone repo: %v", err)
	}

	fmt.Println(gitRepo)

	// read in patterns from config file
	patternFile, err := getFileList(gitRepo, PATTERNFILENAME)
	if err != nil {
		log.Errorf("could not get pattern file: %v", err)
	}

	if len(patternFile) == 0 {
		log.Infof("No pattern file found in git repo. Use default pattern file from sweatShop-analyzer repo.")
		gitRoot, _ := findGitRoot()

		err := getTechsAndPatternsFromFile(filepath.Join(gitRoot, PATTERNFILENAME))
		if err != nil {
			log.Warnf("could not get techs and patterns from file: %v", err)
		}
	}

	// get current commit id for later comparison
	currentCommitID, err := gitRepo.Head()
	if err != nil {
		log.Errorf("could not get current commit id: %v", err)
	}

	log.Println(currentCommitID)

	// Create a new analyzer cache client
	analyzerCacheClient := NewAnalyzerCache(
		goredis.NewClient(&goredis.Options{Addr: redisServer}),
		time.Hour,
	)

	// Try to get cached results
	cached, err := analyzerCacheClient.GetMatchingFiles(repo.Url)
	if err != nil && err != ErrCacheMiss {
		log.Warnf("could not get cached results: %v", err)
	}

	log.Printf("cached: %+v", cached)

	var res []*TechAndPath
	// compared the cached commit id with the current commit id
	if (err != nil && err.Error() == ErrCacheMiss.Error()) || (repo.ForceCompleteAnalysis != nil && *repo.ForceCompleteAnalysis) {

		// If not cached, run initial and complete analysis
		res, err = initialAnalysis(gitRepo)
		if err != nil {
			log.Warnf("could not run initial analysis: %v", err)
		}

	} else if cached != nil && cached.CommitID != currentCommitID.Hash().String() {

		// If cached but commit ids are different, run incremental analysis
		res, err = incrementalAnalysis(gitRepo, cached.CommitID, currentCommitID.Hash().String(), cached.Results)
		if err != nil {
			log.Errorf("could not run incremental analysis: %v", err)
		}

	} else {

		// If cached and commit ids are the same, return cached results
		res = cached.Results
		log.Infof("Using cached results for repo %s: %+v", repo.Url, res)

		// OUTPUT RESULT DATA TO STDOUT FOR NOW
		fmt.Println(res)

		return
	}

	// cache the new commit id and results
	err = analyzerCacheClient.SetMatchingFiles(repo.Url, currentCommitID.Hash().String(), res)
	if err != nil {
		log.Errorf("could not cache results: %v", err)
	}
	log.Infof("Cached results for repo %s: %+v", repo.Url, res)

	// OUTPUT RESULT DATA TO STDOUT FOR NOW
	// WE MIGHT END UP USING REDIS JSON AS A OUTPUT FOMRAT AND ONLY STORE RESULT-IDS IN REDIS STREAMS

	// Create a new analyzer redis json handler
	// go-rejson/v4@v4.1.0 does not support redis/go-redis/v9 (but redis/go-redis/v8)
	// have to create a new client for go-redis/v8 here
	// use redigo conn for simplicity
	analyzerHandler := NewAnalyzerJSONHandlerWithRedigoConn(redisServer)
	defer (*analyzerHandler.conn).Close()

	// Set the results in redis json
	err = analyzerHandler.SetAnalyzerResult(repo, currentCommitID.Hash().String(), res)
	if err != nil {
		log.Errorf("could not set results in redis json: %v", err)
	}

}

func initialAnalysis(gitRepo *git.Repository) ([]*TechAndPath, error) {

	log.Infof("Running initial analysis")

	// init results
	res := make([]*TechAndPath, 0)

	for t, pattern := range techsAndPatterns {
		log.Debugf("Checking for technology %s", t)

		matchingFiles := make([]string, 0)
		// Iterate over the patterns
		for _, p := range pattern {

			mf, err := getFileList(gitRepo, p)
			if err != nil {
				return nil, fmt.Errorf("could not get file list: %v", err)
			}
			log.Debugf("Gathered matching file list: %v", mf)

			matchingFiles = append(matchingFiles, mf...)
		}

		// Get matchingPaths and remove duplicates
		// because one file could fit multiple patterns of the same technology
		matchingPaths := getMatchingPaths(matchingFiles)

		// Append the matching paths to the result
		for _, path := range matchingPaths {
			res = append(res, &TechAndPath{
				Technology: t,
				Path:       path,
			})
		}
	}

	return res, nil
}

func incrementalAnalysis(gitRepo *git.Repository, oldCommitID, newCommitID string, cachedResult []*TechAndPath) ([]*TechAndPath, error) {

	log.Infof("Running incremental analysis")

	// git diff
	patch, err := gitDiff(gitRepo, oldCommitID, newCommitID)
	if err != nil {
		return nil, fmt.Errorf("could not get git diff: %v", err)
	}

	// iterate over git diff output
	for _, fpatch := range patch.FilePatches() {
		log.Tracef("FilePatch: %+v\n", fpatch)

		// check if file is created, deleted, modified or renamed
		// if created or deleted, function returns filesAndStats with only one
		// entry
		// if renamed, function returns filesAndStats with two entries, to
		// remove the old file and add the new file
		// if modified, function returns filesAndStats with zero entries
		filesAndStats := getFilePathAndStatus(fpatch)

		// iterate over files and stats
		for _, v := range filesAndStats {
			file := v.Name
			fstat := v.Stat

			// check if file already in cache
			inCache := checkIfFileInCachedResult(file, cachedResult)

			// If file is new and not in cache, run analysis for file
			if fstat == CREATED && !inCache {
				log.Infof("File %s is new and not in cache", file)

				for t, pattern := range techsAndPatterns {
					log.Debugf("Checking for technology %s", t)

					// Iterate over the patterns
					for _, p := range pattern {

						// Check if file matches pattern
						matches, err := filepath.Match(p, file)
						if err != nil {
							return nil, fmt.Errorf("could not check if file matches pattern: %v", err)
						}

						if matches {
							log.Debugf("File %s matches pattern %s", file, p)

							// Append the matching paths to the result
							cachedResult = append(cachedResult, &TechAndPath{
								Technology: t,
								Path:       file,
							})
						}
					}
				}
			}

			// If file is deleted and in cache, remove file from cache
			if fstat == DELETED && inCache {
				log.Infof("File %s is deleted and in cache", file)

				for i, res := range cachedResult {
					if res.Path == file {
						cachedResult = append(cachedResult[:i], cachedResult[i+1:]...)
					}
				}
			}
		}
	}

	return cachedResult, nil
}

func checkIfFileInCachedResult(file string, cachedResult []*TechAndPath) bool {

	for _, res := range cachedResult {
		if res.Path == file {
			return true
		}
	}

	return false
}
