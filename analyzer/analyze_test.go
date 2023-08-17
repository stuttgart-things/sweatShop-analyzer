package analyzer

import (
	"testing"
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

	for _, tc := range testGetMatchingFilesCases {

		tc.Repo.GetMatchingFiles()

		/*
			res, err := GetMatchingFiles(tc.Repo)
			log.Printf("GetMatchingFiles output: %+v", res)


			assert.Nil(t, err)
			assert.Equal(t, len(tc.Expected.Results), len(res.Results))
			assert.Equal(t, tc.Expected.Results[0].Technology, res.Results[0].Technology)
			assert.Equal(t, tc.Expected.Results[0].Path, res.Results[0].Path)
			assert.NotContains(t, res.Results, "ansible")
		*/
	}
}
