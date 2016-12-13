package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/mitchellh/go-mruby"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")

// Repl encapsulates a series of items used to create a read-evaluate-print
// loop so that end users can manually enter build instructions.
type Repl struct {
	readline *readline.Instance
	mrb      *mruby.Mrb
}

// NewRepl constructs a new Repl.
func NewRepl() (*Repl, error) {
	rl, err := readline.New("kubeshell> ")
	if err != nil {
		return nil, err
	}

	/*
		b, err := builder.NewBuilder(true, []string{})
		if err != nil {
			rl.Close()
			return nil, err
		}
	*/

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	mrb := mruby.NewMrb()
	//defer mrb.Close()

	// Our custom function we'll expose to Ruby. The first return
	// value is what to return from the func and the second is an
	// exception to raise (if any).
	kfunc := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
		pods, err := clientset.Core().Pods("").List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		return nil, nil
	}

	// Lets define a custom class and a class method we can call.
	class := mrb.DefineClass("KubeShell", nil)
	class.DefineClassMethod("pods", kfunc, mruby.ArgsReq(0))

	//return &Repl{readline: rl, builder: b}, nil
	return &Repl{readline: rl, mrb: mrb}, nil
}

// Loop runs the loop. Returns nil on io.EOF, otherwise errors are forwarded.
func (r *Repl) Loop() error {
	defer func() {
		if recover() != nil {
			// interpreter signal or other badness, just abort.
			os.Exit(0)
		}
	}()

	var line string
	for {
		tmp, err := r.readline.Readline()
		if err == io.EOF {
			return nil
		}

		if err != nil && err.Error() == "Interrupt" {
			fmt.Println("You can press ^D or type \"quit\", \"exit\" to exit the shell")
			line = ""
			continue
		}

		if err != nil {
			fmt.Printf("+++ Error %#v\n", err)
			os.Exit(1)
		}

		line += tmp

		switch strings.TrimSpace(line) {
		case "quit":
			fallthrough
		case "exit":
			os.Exit(0)
		}

		//val, err := r.builder.Run(line)
		//line = ""
		//if err != nil {
		//	fmt.Printf("+++ Error: %v\n", err)
		//	continue
		//}

		_, err = r.mrb.LoadString(line)
		line = ""
		if err != nil {
			fmt.Printf("+++ Error: %v\n", err)
			continue
		}

		//fmt.Println(val)
	}
}
