package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/errordeveloper/kubeplay/rubykube"
	mruby "github.com/mitchellh/go-mruby"
)

// Repl encapsulates a series of items used to create a read-evaluate-print
// loop so that end users can play with quebes & pods.
type Repl struct {
	rubykube *rubykube.RubyKube
	readline *readline.Instance
}

// NewRepl constructs a new Repl.
func NewRepl() (*Repl, error) {
	rl, err := readline.New("kubeplay ()> ")
	if err != nil {
		return nil, err
	}

	rk, err := rubykube.NewRubyKube([]string{}, rl)
	if err != nil {
		rl.Close()
		return nil, err
	}

	return &Repl{rubykube: rk, readline: rl}, nil
}

// Loop runs the loop. Returns nil on io.EOF, otherwise errors are forwarded.
func (r *Repl) Loop() error {
	defer func() {
		if err := recover(); err != nil {
			panic(fmt.Errorf("repl.Loop: %v", err))
		}
	}()

	var line string
	var stackKeep int
	//var val *mruby.MrbValue

	parser := mruby.NewParser(r.rubykube.Mrb())
	context := mruby.NewCompileContext(r.rubykube.Mrb())
	context.CaptureErrors(true)

	for {
		tmp, err := r.readline.Readline()
		if err == io.EOF {
			return nil
		}

		if err != nil && err.Error() == "Interrupt" {
			if line != "" {
				r.rubykube.NormalPrompt()
			} else {
				fmt.Println("You can press ^D or type \"quit\", \"exit\" to exit the shell")
			}

			line = ""
			continue
		}

		if err != nil {
			fmt.Printf("+++ Error %#v\n", err)
			os.Exit(1)
		}

		line += tmp + "\n"

		switch strings.TrimSpace(line) {
		case "quit":
			fallthrough
		case "exit":
			os.Exit(0)
		case "help":
			fmt.Println("Please take a look at usage examples\n\t\thttps://github.com/errordeveloper/kubeplay/blob/master/README.md")
		}

		if _, err := parser.Parse(line, context); err != nil {
			r.rubykube.MultiLinePrompt()
			continue
		}

		_, stackKeep, err = r.rubykube.RunCode(parser.GenerateCode(), stackKeep)
		line = ""
		r.rubykube.NormalPrompt()
		if err != nil {
			fmt.Printf("+++ Error: %v\n", err)
			continue
		}

		//if val.String() != "" {
		//	fmt.Println(val)
		//}
	}
}
