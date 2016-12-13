package main

func main() {
	repl, err := NewRepl()
	if err != nil {
		panic(err.Error())
	}
	err = repl.Loop()
	if err != nil {
		panic(err.Error())
	}
}
