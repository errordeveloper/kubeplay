package main

import (
	"github.com/errordeveloper/kubeshell/repl"
)

func main() {
	repl, err := repl.NewRepl()
	if err != nil {
		panic(err.Error())
	}
	err = repl.Loop()
	if err != nil {
		panic(err.Error())
	}
}
