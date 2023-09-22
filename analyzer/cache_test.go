package analyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_GetRevisionMetadata(t *testing.T) {

	redisUtil := initRedisUtilForTesting()
	cache := newAnalyzerCache(redisUtil.Client,
		15*time.Second, // when testing in debug mode, this may time out. Increase if needed.
	)

	// cache miss
	_, err := cache.GetMatchingFiles("my-repo-url")
	assert.Equal(t, ErrCacheMiss, err)
	// populate cache
	res := make([]*TechAndPath, 0)
	res = append(res, &TechAndPath{Technology: "my-tech", Path: "my-path"})
	err = cache.SetMatchingFiles("my-repo-url", "my-commit-id", res)
	assert.NoError(t, err)
	// cache miss
	_, err = cache.GetMatchingFiles("other-repo-url")
	assert.Equal(t, ErrCacheMiss, err)
	// cache hit
	value, err := cache.GetMatchingFiles("my-repo-url")
	assert.NoError(t, err)
	assert.Equal(t, &MatchingFilesValue{
		CommitID: "my-commit-id",
		Results:  res,
	}, value)
	// cleanup
	err = cache.SetItem(matchingFilesKey("my-repo-url"), nil, cache.expiration, true)
	assert.NoError(t, err)
}
