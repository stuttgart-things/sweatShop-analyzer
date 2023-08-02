package analyzer

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

const PATTERNFILENAME = "yacht-analyze.yaml"

var techsAndPatterns = make(map[string][]string)

// getTechsAndPatternsFromFile returns a map of technologies and their patterns
func getTechsAndPatternsFromFile(path string) error {
	log.Infof("Get techs and patterns from file %s", path)

	// If file does not exist, return empty map
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("Path does not exist")
			return err
		} else {
			log.Debugf("Error occurred while checking path: %v", err)
			return err
		}
	}

	// If file exists, read in file
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Debugf("Error reading YAML file: %v", err)
		return err
	}

	// Parse file into map
	err = yaml.Unmarshal(yamlFile, &techsAndPatterns)
	if err != nil {
		log.Debugf("Error parsing YAML file: %v", err)
		return err
	}

	return nil
}

// TODO: how to make sure the techs are consistent as key?
