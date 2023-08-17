/*
Copyright © 2023 PATRICK HERMANN patrick.hermann@sva.de
*/

package stream

import (
	"reflect"
	"testing"

	"github.com/stuttgart-things/sweatShop-analyzer/analyzer"
)

var testCases_isValidAnalyzerInput = []struct {
	Values   map[string]interface{}
	Expected *analyzer.Repository
}{
	{
		Values:   map[string]interface{}{},
		Expected: nil,
	},
	{
		Values: map[string]interface{}{
			"name":     "test invalid url",
			"url":      "deeply.invalid.url",
			"revision": "main",
		},
		Expected: nil,
	},
	{
		Values: map[string]interface{}{
			"name":     "test invalid auth",
			"url":      "https://codehub.sva.de/Lab/stuttgart-things/yacht/yacht-analyze.git",
			"revision": "main",
			"username": "test",
			"password": "test",
		},
		Expected: nil,
	},
	{
		Values: map[string]interface{}{
			"name": "test empty revision",
			"url":  "https://github.com/geerlingguy/ansible-role-gitlab",
		},
		Expected: nil,
	},
	{
		Values: map[string]interface{}{
			"name":     "test valid input",
			"url":      "https://github.com/fluxcd/flux2",
			"revision": "main",
		},
		Expected: &analyzer.Repository{
			Name:     "test valid input",
			Url:      "https://github.com/fluxcd/flux2",
			Revision: "main",
		},
	},
}

func Test_isValidAnalyzerInput(t *testing.T) {

	for _, tc := range testCases_isValidAnalyzerInput {
		actual := buildValidRepository(tc.Values)
		if reflect.DeepEqual(actual, tc.Expected) != true {
			t.Errorf("isValidAnalyzerInput(%v): expected %v, actual %v", tc.Values, tc.Expected, actual)
		}
	}
}
