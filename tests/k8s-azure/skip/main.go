package main

type Rankings struct {
	Keyword  string
	GetCount uint32
	Engine   string
	Locale   string
	Mobile   bool
}

func main() {
	result, _ := extractSkipKey("b")
	generateSkipFile("./", result)
}
