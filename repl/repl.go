package repl

import (
	"flag"
	"fmt"
	"io"
	"os"
	_ "runtime/debug"
	"strings"

	"github.com/chzyer/readline"
	"github.com/errordeveloper/kubeshell/rubykube"
	"github.com/mitchellh/go-mruby"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")

// Repl encapsulates a series of items used to create a read-evaluate-print
// loop so that end users can manually enter build instructions.
type Repl struct {
	readline *readline.Instance
	mrb      *mruby.Mrb
	rk       *rubykube.RubyKube
}

// NewRepl constructs a new Repl.
func NewRepl() (*Repl, error) {
	rl, err := readline.New("kubeshell> ")
	if err != nil {
		return nil, err
	}

	rk, err := rubykube.NewRubyKube([]string{})
	if err != nil {
		rl.Close()
		return nil, err
	}

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	_, err = kubernetes.NewForConfig(config) // clientset
	if err != nil {
		panic(err.Error())
	}
	mrb := mruby.NewMrb()
	//defer mrb.Close()

	/*

		type Method struct {
			Func    mruby.Func
			ArgsReq int
		}

		type containerClassVars struct {
			Container v1.Container
		}

		ctx := containerClassVars{}

		containerClassMethods := map[string]Method{
			"count_pods": Method{
				Func: func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
					pods, err := clientset.Core().Pods("").List(v1.ListOptions{})
					if err != nil {
						panic(err.Error())
					}
					fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
					return nil, nil
				},
				ArgsReq: 0,
			},
			//"decode_test": Method{
			//	Func: func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			//		var logData interface{}
			//		args := m.GetArgs()
			//		if err := mruby.Decode(&logData, args[0]); err != nil {
			//			panic(err.Error())
			//		}

			//		fmt.Printf("Result: %+v\n", logData)

			//		return nil, nil
			//	},
			//	ArgsReq: 1,
			//},
			"initialize": Method{
				Func: func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
					args := m.GetArgs()
					ctx.Container = v1.Container{Image: args[0].String()}

					imageParts := strings.Split(strings.Split(ctx.Container.Image, ":")[0], "/")
					ctx.Container.Name = imageParts[len(imageParts)-1]
					if len(args) > 1 {
						ctx.Container.Name = args[1].String()
					}

					fmt.Printf("%#v\n", ctx.Container)

					return nil, nil
				},
				ArgsReq: 1,
			},
		}

		containerClass := mrb.DefineClass("Container", nil)
		for k, v := range containerClassMethods {
			containerClass.DefineClassMethod(k, v.Func, mruby.ArgsReq(v.ArgsReq))
		}

		//return &Repl{readline: rl, builder: b}, nil
	*/
	return &Repl{readline: rl, mrb: mrb, rk: rk}, nil
}

// Loop runs the loop. Returns nil on io.EOF, otherwise errors are forwarded.
func (r *Repl) Loop() error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			// interpreter signal or other badness, just abort.
			//debug.PrintStack()
			os.Exit(3)
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

		val, err := r.rk.Run(line)
		line = ""
		if err != nil {
			fmt.Printf("+++ Error: %v\n", err)
			continue
		}

		//_, err = r.mrb.LoadString(line)
		//line = ""
		//if err != nil {
		//	fmt.Printf("+++ Error: %v\n", err)
		//	continue
		//}

		fmt.Println(val)
	}
}
