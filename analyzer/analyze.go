package analyzer

import (
	"fmt"
	"path/filepath"

	sthingsBase "github.com/stuttgart-things/sthingsBase"
)

type Repository struct {
	Name     string
	Url      string
	Revision string
	Username string
	Password string
	Insecure bool
}

var (
	log         = sthingsBase.StdOutFileLogger(logfilePath, "2006-01-02 15:04:05", 50, 3, 28)
	logfilePath = "/tmp/sweatShop-analyze.log"
)

func GetMatchingFiles(repo Repository) {
	log.Println("sweatShop-analyze started")
	log.Println("GetMatchingFiles for repo:", repo)

	// Clone the repo into memory for later use
	gitRepo, err := gitCloneRevision(repo)

	if err != nil {
		log.Errorf("could not clone repo: %w", err)
	}

	fmt.Println(gitRepo)

	// read in patterns from config file
	patternFile, err := getFileList(gitRepo, PATTERNFILENAME)
	if err != nil {
		log.Errorf("could not get pattern file: %w", err)
	}

	if len(patternFile) == 0 {
		log.Infof("No pattern file found in git repo. Use default pattern file from yacht-analyze repo.")
		gitRoot, _ := findGitRoot()

		err := getTechsAndPatternsFromFile(filepath.Join(gitRoot, PATTERNFILENAME))
		if err != nil {
			log.Errorf("could not get techs and patterns from file: %w", err)
		}
	}

}
