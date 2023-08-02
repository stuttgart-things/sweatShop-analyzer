/*
Copyright © 2023 XIAOMIN LAI
*/

package analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
)

func findGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("git root not found")
}

func getMatchingPaths(mf []string) []string {

	res := make([]string, 0)

	for _, v := range mf {
		vpath := filepath.Dir(v)
		if slices.Contains(res, vpath) {
			continue
		}
		res = append(res, vpath)
	}

	return res
}
