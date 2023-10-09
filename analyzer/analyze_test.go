package analyzer

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/nitishm/go-rejson/v4"
	"github.com/stretchr/testify/assert"
)

var testGetMatchingFilesCases = []struct {
	Repo           *Repository
	CachedCommitID string
	Expected       []*TechAndPath
}{
	{
		Repo: &Repository{
			Url:      "https://github.com/fluxcd/flux2",
			Revision: "main",
		},
		CachedCommitID: "1daa7a8aa4e79fd3d6d788b628c85942e340cbc7",
		Expected: []*TechAndPath{
			{
				Technology: "golang",
				Path:       ".",
			},
			{
				Technology: "docker",
				Path:       ".",
			},
		},
	},
	{
		Repo: &Repository{
			Url:      "https://github.com/geerlingguy/ansible-role-gitlab",
			Revision: "master",
		},
		CachedCommitID: "025a0c517fe3d4eda76a6efe53c4f81fbe8d9ec7",
		Expected: []*TechAndPath{
			{
				Technology: "ansible-role",
				Path:       "meta",
			},
			{
				Technology: "ansible-role",
				Path:       "tasks",
			},
		},
	},
	{
		Repo: &Repository{
			Url:      "https://github.com/aws-samples/eks-gitops-crossplane-argocd",
			Revision: "main",
		},
		CachedCommitID: "6ca884922959c9d7c44287875c41b6218bb32185",
		Expected: []*TechAndPath{
			{
				Technology: "helm",
				Path:       "crossplane-complete",
			},
			{
				Technology: "helm",
				Path:       "workload-apps",
			},
		},
	},
}

func TestGetMatchingFiles(t *testing.T) {

	redisClient, _ := redismock.NewClientMock()
	cache := &AnalyzerCacheMock{
		AnalyzerCache: *NewAnalyzerCache(
			redisClient,
			15*time.Second, // when testing in debug mode, this may time out. Increase if needed.
		),
	}
	cache.MockedSetMatchingFiles = func(repoURL, commitId string, res []*TechAndPath) error {
		return nil
	}
	h := new(AnalyzerJSONHandlerMock)
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClientWithContext(context.Background(), redisClient)
	h.AnalyzerJSONHandler = *NewAnalyzerJSONHandler(rh)
	h.MockedSetAnalyzerResult = func(repoURL *Repository, commitId string, res []*TechAndPath) error {
		return nil
	}

	// run initial analysis
	for _, tc := range testGetMatchingFilesCases {

		// change GetMatchingFiles to return cache miss
		cache.MockedGetMatchingFiles = func(repoURL string) (*MatchingFilesValue, error) {
			return nil, ErrCacheMiss
		}

		err := tc.Repo.GetMatchingFiles(cache, h)
		assert.Nil(t, err)

		// change GetMatchingFiles to return cached results
		cache.MockedGetMatchingFiles = func(repoURL string) (*MatchingFilesValue, error) {
			return &MatchingFilesValue{
				CommitID: tc.CachedCommitID,
				Results:  tc.Expected,
			}, nil
		}

		// use cached results
		err = tc.Repo.GetMatchingFiles(cache, h)
		assert.Nil(t, err)
	}

}
