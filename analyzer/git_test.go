package analyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_gitDiff(t *testing.T) {

	// Create a new redis cache
	redisUtil := initRedisUtilForTesting()
	cache := newAnalyzerCache(redisUtil.Client,
		30*time.Second, // when testing in debug mode, this may time out. Increase if needed.
	)

	// Create a new repo
	repo := &Repository{
		Name:     "testing-repo",
		Url:      "https://github.com/go-git/go-git",
		Revision: "wasm",
	}

	gitRepo, _ := gitCloneRevision(repo)
	currentCommitID, _ := gitRepo.Head()

	// populate cache
	err := cache.SetMatchingFiles(repo.Url, currentCommitID.Hash().String(), nil)
	assert.NoError(t, err)

	// change the revision
	repo.Revision = "master"
	gitRepo, _ = gitCloneRevision(repo)
	currentCommitID, _ = gitRepo.Head()

	cached, _ := cache.GetMatchingFiles(repo.Url)

	diff, err := gitDiff(gitRepo, cached.CommitID, currentCommitID.Hash().String())
	assert.NoError(t, err)
	assert.NotNil(t, diff)

	// cleanup
	err = cache.SetItem(analyzerResultKey(repo.Url), nil, time.Second, true)
	assert.NoError(t, err)
}
