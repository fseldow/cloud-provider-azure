package main

import "fmt"

func main() {
	a, _ := readSkipFile("skip.log.json")
	fmt.Println(len(a))
}
