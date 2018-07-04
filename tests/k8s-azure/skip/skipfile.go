package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var skipListFromReport []string
var failListFromReport []string

// SkipTest is the structure for skip.log.json
type SkipTest struct {
	Name    string   `json:"Name"`
	Comment string   `json:"Comment"`
	Subtest []string `json:"Subtest"`
}

func extractSkipKey(filename string) (result []SkipTest, err error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)
	text := string(bytes)
	lines := strings.Split(text, "\r\n")

	var testClass SkipTest

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
	// To avoid no subtest
	for _, test := range result {
		if len(test.Subtest) == 0 {
			test.Subtest = append(test.Subtest, test.Name)
		}
	}
	return
}

func generateSkipFile(fileDir string, tests []SkipTest) (err error) {
	logfile := fileDir + "skip.log.json"
	skipfile := fileDir + "skip.txt"
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

func readSkipFile(file string) (result []SkipTest, err error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &result)
	return
}

func UpdateSkipAndFailListFromReport(dirPath string) {
	EmptyList()
	filepath.Walk(dirPath, visitJunitReport)
}

func EmptyList() {
	skipListFromReport = make([]string, 0)
	failListFromReport = make([]string, 0)
}

func visitJunitReport(fp string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err) // can't walk here,
		return nil       // but continue walking elsewhere
	}
	if fi.IsDir() {
		return nil // not a file.  ignore.
	}
	matched, err := filepath.Match("junit_*.xml", fi.Name())
	if err != nil {
		fmt.Println(err) // malformed pattern
		return err       // this is fatal.
	}
	if matched {
		updateSkipListFromJunit(fp)
	}
	return nil
}

func updateSkipListFromJunit(file string) (err error) {
	xmlFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var suite TestSuite
	xml.Unmarshal(byteValue, &suite)
	for _, tc := range suite.TestCases {
		if tc.Time == 0 {
			skipListFromReport = append(skipListFromReport, tc.Name)
		}
		if len(tc.Failures) > 0 {
			failListFromReport = append(failListFromReport, tc.Name)
		}
	}
	return
}

func Match(skipFromFile []SkipTest, skipFromReport []string) bool {
	sort.Strings(skipFromReport)
	sort.Slice(skipFromFile, func(i, j int) bool { return skipFromFile[i].Name < skipFromFile[j].Name })
	for _, st := range skipFromFile {
		sort.Strings(st.Subtest)
	}

	var p1, p2 int = 0, 0
	var subPoint int = 0
	for p1 < len(skipFromReport) && p2 < len(skipFromFile) {
		if strings.Contains(skipFromReport[p1], skipFromFile[p2].Name) {
			// check subtests
			for subPoint < len(skipFromFile[p2].Subtest) {
				if skipFromReport[p1] == skipFromFile[p2].Subtest[subPoint] {
					p1++
					subPoint++
				} else if skipFromReport[p1] < skipFromFile[p2].Subtest[subPoint] {
					//TODO leak
					p1++
				} else {
					//TODO subtest not exist
					subPoint++
				}
			}
		} else if skipFromReport[p1] < skipFromFile[p2].Name {
			p1++
		} else {
			if subPoint < len(skipFromFile[p2].Subtest) {
				//TODO certain subtests not exist
			}
			subPoint = 0
			p2++
		}
	}

	return true
}
