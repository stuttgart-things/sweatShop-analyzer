package analyzer

import (
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

type AnalyzerCacheMock struct {
	AnalyzerCache
	MockedGetMatchingFiles func(repoURL string) (*MatchingFilesValue, error)
	MockedSetMatchingFiles func(repoURL, commitId string, res []*TechAndPath) error
}

func (acm *AnalyzerCacheMock) GetMatchingFiles(repoURL string) (*MatchingFilesValue, error) {
	return acm.MockedGetMatchingFiles(repoURL)
}

func (acm *AnalyzerCacheMock) SetMatchingFiles(repoURL, commitId string, res []*TechAndPath) error {
	return acm.MockedSetMatchingFiles(repoURL, commitId, res)
}

var tcCache = []*TechAndPath{
	{
		Technology: "my-tech",
		Path:       "my-path",
	},
}

func TestCache_GetMatchingFiles(t *testing.T) {

	redisClient, _ := redismock.NewClientMock()
	cache := new(AnalyzerCacheMock)
	cache.AnalyzerCache = *NewAnalyzerCache(
		redisClient,
		15*time.Second, // when testing in debug mode, this may time out. Increase if needed.
	)

	// cache miss
	cache.MockedGetMatchingFiles = func(repoURL string) (*MatchingFilesValue, error) {
		return nil, ErrCacheMiss
	}
	_, err := cache.GetMatchingFiles("my-repo-url")
	assert.Equal(t, ErrCacheMiss, err)

	// populate cache
	cache.MockedSetMatchingFiles = func(repoURL, commitId string, res []*TechAndPath) error {
		return nil
	}
	err = cache.SetMatchingFiles("my-repo-url", "my-commit-id", tcCache)
	assert.NoError(t, err)

	// cache hit
	cache.MockedGetMatchingFiles = func(repoURL string) (*MatchingFilesValue, error) {
		return &MatchingFilesValue{
			CommitID: "my-commit-id",
			Results:  tcCache,
		}, nil
	}
	value, err := cache.GetMatchingFiles("my-repo-url")
	assert.NoError(t, err)
	assert.Equal(t, &MatchingFilesValue{
		CommitID: "my-commit-id",
		Results:  tcCache,
	}, value)

}
