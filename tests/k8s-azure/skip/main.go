package main

func main() {
	//generateSkipFile("./")
	reportlist := GetReportList("./t-xinhli-new3/")
	skipDescriptions, _ := readSkipFile("skip.log.json")
	result_list := newSkips(skipDescriptions, reportlist)
	newSkip, focuslist := checkTest(skipDescriptions, result_list, false)
}
