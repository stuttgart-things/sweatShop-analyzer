package analyzer

import (
	"testing"

	goredis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func Test_AnalyzerResultValueWithGoRedisClient(t *testing.T) {

	h := NewAnalyzerJSONHandlerWithGoRedisClient(
		goredis.NewClient(&goredis.Options{Addr: redisServer}),
	)

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
	h.SetItem(analyzerResultKey(testValue.Repo.Url), nil, true)
}

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
	h.SetItem(analyzerResultKey(testValue.Repo.Url), nil, true)
}

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
