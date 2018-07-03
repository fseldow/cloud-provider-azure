package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type TestFramework struct {
	Name    string   `json:"Name"`
	Comment string   `json:"Comment"`
	Subtest []string `json:"Subtest"`
}

func extractSkipKey(filename string) (result []TestFramework, err error) {
	file, err := os.Open("../skip.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)
	text := string(bytes)
	lines := strings.Split(text, "\r\n")

	var testClass TestFramework

	for _, line := range lines {
		if line == "" {
			testClass.Comment = ""

		} else if line[0] != '#' {
			// Name
			testClass.Name = line
			result = append(result, testClass)

			var myslice []string
			testClass.Subtest = myslice

		} else if line[1] == ' ' || line[1] == '#' {
			// comment
			testClass.Comment += line[2 : len(line)-1]

		} else {
			// sub test
			testClass.Subtest = append(testClass.Subtest, line[1:len(line)-1])

		}
	}

	return
}

func generateSkipFile(filePath string, tests []TestFramework) (err error) {
	logfile := filePath + "skip.log.json"
	skipfile := filePath + "skip.txt"
	rankingsJSON, _ := json.Marshal(tests)
	ioutil.WriteFile(logfile, rankingsJSON, 0644)
	var file *os.File

	if file, err = os.Create(skipfile); err != nil {
		return
	}

	defer file.Close()

	for _, test := range tests {
		_, err := file.WriteString(strings.TrimSpace("# "+test.Comment) + "\n")
		_, err = file.WriteString(strings.TrimSpace(test.Name) + "\n")
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	return
}
