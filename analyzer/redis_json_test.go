package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-redis/redismock/v9"
	"github.com/nitishm/go-rejson/v4"
)

type AnalyzerJSONHandlerMock struct {
	AnalyzerJSONHandler
	MockedGetAnalyzerResult func(repoURL string) (*AnalyzerResultValue, error)
	MockedSetAnalyzerResult func(repoURL *Repository, commitId string, res []*TechAndPath) error
}

func (acm *AnalyzerJSONHandlerMock) GetAnalyzerResult(repoURL string) (*AnalyzerResultValue, error) {
	return acm.MockedGetAnalyzerResult(repoURL)
}

func (acm *AnalyzerJSONHandlerMock) SetAnalyzerResult(repoURL *Repository, commitId string, res []*TechAndPath) error {
	return acm.MockedSetAnalyzerResult(repoURL, commitId, res)
}

func Test_AnalyzerResultValueWithGoRedisClient(t *testing.T) {

	redisClient, _ := redismock.NewClientMock()
	h := new(AnalyzerJSONHandlerMock)
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClientWithContext(context.Background(), redisClient)
	h.AnalyzerJSONHandler = *NewAnalyzerJSONHandler(rh)

	// json miss
	h.MockedGetAnalyzerResult = func(repoURL string) (*AnalyzerResultValue, error) {
		return nil, fmt.Errorf(ErrJSONMissWithGoRedisClient)
	}
	_, err := h.GetAnalyzerResult("my-repo-url")
	assert.Equal(t, ErrJSONMissWithGoRedisClient, err.Error())

	// populate json
	h.MockedSetAnalyzerResult = func(repoURL *Repository, commitId string, res []*TechAndPath) error {
		return nil
	}
	err = h.SetAnalyzerResult(testValue.Repo, testValue.Commit, testValue.Results)
	assert.NoError(t, err)

	// json hit
	h.MockedGetAnalyzerResult = func(repoURL string) (*AnalyzerResultValue, error) {
		return testValue, nil
	}
	value, err := h.GetAnalyzerResult("my-repo-url")
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)
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
