package main

type Tests struct {
	Suite TestSuite  `xml:"testsuite"`
	Tests []TestCase `xml:"testcase"`
}

type TestSuite struct {
	count   int     `xml:"tests"`
	failure int     `xml:"failure"`
	time    float64 `xml:"time"`
}

type TestCase struct {
	name      string  `xml:"name"`
	classname string  `xml:"classname"`
	time      float64 `xml:"time"`
}
