package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	redisutil "github.com/stuttgart-things/sweatShop-analyzer/utils/redis"
)

func initRedisUtilForTesting() *redisutil.Redis {
	r := redisutil.NewRedisWithClient(
		"localhost",
		6379,
		"Atlan7is",
	)
	r.SetJSONHandler()
	return r
}

func Test_AnalyzerResultValueWithGoRedisClient(t *testing.T) {

	redisUtil := initRedisUtilForTesting()
	h := newAnalyzerJSONHandler(redisUtil.JSONHandler)

	// json miss
	_, err := h.GetAnalyzerResult("my-repo-url")
	assert.Equal(t, ErrJSONMissWithGoRedisClient, err.Error())
	// populate json
	err = h.SetAnalyzerResult(testValue.Repo, testValue.Commit, testValue.Results)
	assert.NoError(t, err)
	// json miss
	_, err = h.GetAnalyzerResult("other-repo-url")
	assert.Equal(t, ErrJSONMissWithGoRedisClient, err.Error())
	// json hit
	value, err := h.GetAnalyzerResult("my-repo-url")
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)
	// cleanup
	err = h.SetItem(analyzerResultKey(testValue.Repo.Url), nil, true)
	assert.NoError(t, err)
}

/*
func Test_AnalyzerResultValueWithRedigoConn(t *testing.T) {

	h := NewAnalyzerJSONHandlerWithRedigoConn("localhost:6379")

	// json miss
	_, err := h.GetAnalyzerResult("my-repo-url")
	assert.Equal(t, ErrJSONMissWithRedigoConn, err)
	// populate json
	err = h.SetAnalyzerResult(testValue.Repo, testValue.Commit, testValue.Results)
	assert.NoError(t, err)
	// json miss
	_, err = h.GetAnalyzerResult("other-repo-url")
	assert.Equal(t, ErrJSONMissWithRedigoConn, err)
	// json hit
	value, err := h.GetAnalyzerResult("my-repo-url")
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)
	// cleanup
	err = h.SetItem(analyzerResultKey(testValue.Repo.Url), nil, true)
	assert.NoError(t, err)
}
*/

var testValue = &AnalyzerResultValue{
	Repo: &Repository{
		Url: "my-repo-url",
	},
	Commit: "my-commit-id",
	Results: []*TechAndPath{
		{
			Technology: "go",
			Path:       ".",
		},
		{
			Technology: "Dockefile",
			Path:       ".",
		},
	},
}
