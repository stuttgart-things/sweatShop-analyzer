package main

import (
	"github.com/stuttgart-things/sweatShop-analyze/analyzer"
)

func main() {

	repo := analyzer.Repository{
		"stuttgart-things",
		"https://github.com/stuttgart-things/stuttgart-things.git",
		"main",
		"",
		"",
		false}
	analyzer.GetMatchingFiles(repo)

}
