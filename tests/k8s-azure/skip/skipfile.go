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

const (
	PassStatus = "pass"
	SkipStatus = "skip"
	FailStatus = "fail"
)

type SkipSuite struct {
	Descriptions []Description
}

type Report struct {
	Name   string `json:"Name"`             // test name
	Status string `json:"Status,omitempty"` // 1->pass, 0->skip, -1->fail
}

// SkipTest is the structure for skip.log.json
type Description struct {
	Name    string   `json:"Descirbe"`
	Comment string   `json:"Comment"`
	Subtest []Report `json:"Subtest"`
}

func extractSkipKey(filename string) (result []Description, err error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)
	text := string(bytes)
	lines := strings.Split(text, "\r\n")

	var testClass Description

	for _, line := range lines {
		if line == "" {
			testClass.Comment = ""
		} else if line[0] != '#' {
			// Name
			testClass.Name = line
			result = append(result, testClass)

			testClass.Subtest = make([]Report, 0)
		} else if line[1] == ' ' || line[1] == '#' {
			// comment
			testClass.Comment += line[2:]
		} else {
			// sub test
			testClass.Subtest = append(testClass.Subtest, Report{line[1:], SkipStatus})
		}
	}
	// To avoid no subtest
	for i := range result {
		if result[i].Subtest == nil {
			result[i].Subtest = make([]Report, 0)
		}
		if len(result[i].Subtest) == 0 {
			result[i].Subtest = append(result[i].Subtest, Report{result[i].Name, SkipStatus})
		}
	}
	return
}

func generateSkipFile(fileDir string) (err error) {
	logfile := fileDir + "skip.log.json"
	skipfile := fileDir + "skip.txt"

	descrips, err := extractSkipKey("../skip.txt")
	if err != nil {
		return err
	}
	suiteJSON, _ := json.MarshalIndent(descrips, "", "  ")

	ioutil.WriteFile(logfile, suiteJSON, 0644)
	var file *os.File
	if file, err = os.Create(skipfile); err != nil {
		return
	}
	defer file.Close()

	for _, test := range descrips {
		_, err := file.WriteString(strings.TrimSpace("# "+test.Comment) + "\r\n")
		_, err = file.WriteString(strings.TrimSpace(test.Name) + "\r\n")
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	return
}

func readSkipFile(file string) (result []Description, err error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &result)
	return
}

func GetReportList(dirPath string) (reportList []Report) {
	reportList = make([]Report, 0)

	updateReportFromJunit := func(file string) (err error) {
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
				reportList = append(reportList, Report{tc.Name, SkipStatus})
			} else if len(tc.Failures) > 0 {
				reportList = append(reportList, Report{tc.Name, FailStatus})
			} else {
				reportList = append(reportList, Report{tc.Name, PassStatus})
			}
		}
		return
	}

	visitJunitReport := func(fp string, fi os.FileInfo, err error) error {
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
			updateReportFromJunit(fp)
		}
		return nil
	}
	filepath.Walk(dirPath, visitJunitReport)
	return
}

func contains(s []Report, e string) bool {
	for _, a := range s {
		if a.Name == e {
			return true
		}
	}
	return false
}

func newSkips(skipFromFile []Description, reportList []Report) (result []Description) {
	for i := 0; i < len(skipFromFile); i++ {
		temp := Description{
			skipFromFile[i].Name,
			skipFromFile[i].Comment,
			make([]Report, 0),
		}
		for j := 0; j < len(reportList); j++ {
			if strings.Contains(reportList[j].Name, skipFromFile[i].Name) {
				if !contains(temp.Subtest, reportList[j].Name) {
					temp.Subtest = append(temp.Subtest, reportList[j])
				}
			}
		}
		result = append(result, temp)
	}
	suiteJSON, _ := json.MarshalIndent(result, "", "  ")
	ioutil.WriteFile("test.json", suiteJSON, 0644)
	return
}

func checkTest(skips []Description, reports []Description, ifSecond bool) (result []Description, focus []string) {
	var index int
	for index = range skips {
		sort.Slice(skips[index].Subtest, func(i, j int) bool { return skips[index].Subtest[i].Name < skips[index].Subtest[j].Name })
		sort.Slice(reports[index].Subtest, func(i, j int) bool { return reports[index].Subtest[i].Name < reports[index].Subtest[j].Name })
	}
	for index = range skips {
		skipSub := skips[index].Subtest
		var i int
		temp := Description{
			skips[index].Name,
			skips[index].Comment,
			make([]Report, 0),
		}
		toSplit := false
		for _, report := range reports[index].Subtest {
			if i >= len(skipSub) || skipSub[i].Name > report.Name {
				if report.Status == FailStatus {
					temp.Subtest = append(temp.Subtest, report)
				} else if report.Status == PassStatus {
					toSplit = true
				} else {
					//temp.Subtest = append(temp.Subtest, report)
					if !ifSecond {
						focus = append(focus, report.Name)
					} else {
						temp.Subtest = append(temp.Subtest, report)
					}
				}
			} else if skipSub[i].Name < report.Name {
				i++
			} else {
				if report.Status == PassStatus {
					toSplit = true
				} else {
					temp.Subtest = append(temp.Subtest, report)
					i++
				}
			}
		}
		if !toSplit {
			result = append(result, temp)
		} else {
			for _, test := range temp.Subtest {
				description := Description{
					test.Name,
					temp.Comment,
					[]Report{test},
				}
				result = append(result, description)
			}
		}
	}
	return
}
