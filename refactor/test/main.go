package main

import (
	"fmt"
	"io/ioutil"

	"github.com/bblfsh/go-driver/refactor"
)

const (
	before = "../data/before"
	after  = "../data/after"
	test   = "../data/example.go"
)

func main() {
	refactor, err := refactor.NewRefactor(before, after)
	if err != nil {
		panic(err)
	}

	testData, err := ioutil.ReadFile(test)
	if err != nil {
		panic(err)
	}

	if err := refactor.Prepare(); err != nil {
		panic(err)
	}

	code, err := refactor.Apply(string(testData))
	if err != nil {
		panic(err)
	}

	fmt.Println(code)
}
