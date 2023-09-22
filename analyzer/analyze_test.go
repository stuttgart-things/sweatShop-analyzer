package analyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testGetMatchingFilesCases = []struct {
	Repo     *Repository
	Expected []*TechAndPath
}{
	{
		Repo: &Repository{
			Url:      "https://github.com/fluxcd/flux2",
			Revision: "main",
		},
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

	redisUtil := initRedisUtilForTesting()
	redisUtil.SetJSONHandler()
	cache := newAnalyzerCache(redisUtil.Client, 30)
	h := newAnalyzerJSONHandler(redisUtil.JSONHandler)

	// run initial analysis
	for _, tc := range testGetMatchingFilesCases {
		err := tc.Repo.GetMatchingFiles(redisUtil)
		assert.Nil(t, err)
	}

	// use cached results
	for _, tc := range testGetMatchingFilesCases {
		err := tc.Repo.GetMatchingFiles(redisUtil)
		assert.Nil(t, err)
	}

	// cleanup
	for _, tc := range testGetMatchingFilesCases {
		err := cache.SetItem(matchingFilesKey(tc.Repo.Url), nil, time.Second, true)
		assert.Nil(t, err)

		err = h.SetItem(analyzerResultKey(tc.Repo.Url), nil, true)
		assert.Nil(t, err)
	}
}
