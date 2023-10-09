package analyzer

import (
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func Test_gitDiff(t *testing.T) {

	redisClient, _ := redismock.NewClientMock()
	cache := new(AnalyzerCacheMock)
	cache.AnalyzerCache = *NewAnalyzerCache(
		redisClient,
		15*time.Second, // when testing in debug mode, this may time out. Increase if needed.
	)

	// Create a new repo
	repo := &Repository{
		Name:     "testing-repo",
		Url:      "https://github.com/go-git/go-git",
		Revision: "wasm",
	}

	gitRepo, _ := gitCloneRevision(repo)
	firstCommitID, _ := gitRepo.Head()

	// populate cache
	cache.MockedSetMatchingFiles = func(repoURL, commitId string, res []*TechAndPath) error {
		return nil
	}
	err := cache.SetMatchingFiles(repo.Url, firstCommitID.Hash().String(), nil)
	assert.NoError(t, err)
	cache.MockedGetMatchingFiles = func(repoURL string) (*MatchingFilesValue, error) {
		return &MatchingFilesValue{
			CommitID: firstCommitID.Hash().String(),
		}, nil
	}

	// change the revision
	repo.Revision = "master"
	gitRepo, _ = gitCloneRevision(repo)
	secondCommitID, _ := gitRepo.Head()

	cached, _ := cache.GetMatchingFiles(repo.Url)

	diff, err := gitDiff(gitRepo, cached.CommitID, secondCommitID.Hash().String())
	assert.NoError(t, err)
	assert.NotNil(t, diff)
}
