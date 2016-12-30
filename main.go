package main

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/repl"
)

func main() {
	repl, err := repl.NewRepl()
	if err != nil {
		panic(fmt.Errorf("repl.NewRepl: %v", err))
	}
	err = repl.Loop()
	if err != nil {
		panic(fmt.Errorf("repl.Loop: %v", err))
	}
}
