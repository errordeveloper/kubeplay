package rubykube

import (
	"flag"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/erikh/box/builder/signal"
	"github.com/erikh/box/log"
	mruby "github.com/mitchellh/go-mruby"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

type RubyKube struct {
	mrb       *mruby.Mrb
	clientset *kubernetes.Clientset
	classes   Classes
	readline  *readline.Instance
	state     *CurrentState
}

type Classes struct {
	Root     *mruby.Class
	Pods     *podsClass
	Pod      *podClass
	PodMaker *podMaker
}

type CurrentState struct {
	Namespace string
	Cluster   string
	Context   string
}

func keep(omitFuncs []string, name string) bool {
	for _, fun := range omitFuncs {
		if name == fun {
			return false
		}
	}
	return true
}

// NewRubyKube may return an error on mruby or k8s.io/client-go issues.
func NewRubyKube(omitFuncs []string, rl *readline.Instance) (*RubyKube, error) {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(fmt.Errorf("clientcmd.BuildConfigFromFlags: %v", err))
	}

	rk := &RubyKube{mrb: mruby.NewMrb(), readline: rl, state: &CurrentState{}}

	rk.mrb.DisableGC()

	for name, def := range verbJumpTable {
		if keep(omitFuncs, name) {
			rk.AddVerb(name, def.verbFunc, def.argSpec)
		}
	}

	for name, def := range funcJumpTable {
		if keep(omitFuncs, name) {
			inner := def.fun
			fn := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return inner(rk, m, self)
			}

			rk.mrb.TopSelf().SingletonClass().DefineMethod(name, fn, def.argSpec)
		}
	}

	signal.SetSignal(nil)

	rk.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("kubernetes.NewForConfig: %v", err))
	}

	rk.classes = Classes{Root: rk.mrb.DefineClass("RubyKube", nil)}
	rk.classes.Pods = newPodsClass(rk)
	rk.classes.Pod = newPodClass(rk)
	rk.classes.PodMaker = newPodMakerClass(rk)

	rk.SetNamespace("*")
	if err := rk.applyPatches(); err != nil {
		return nil, err
	}

	return rk, nil
}

// AddVerb adds a function to the mruby dispatch as well as adding hooks around
// the call to ensure containers are committed and intermediate layers are
// cleared.
func (rk *RubyKube) AddVerb(name string, fn verbFunc, args mruby.ArgSpec) {
	hookFunc := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
		args := m.GetArgs()
		strArgs := extractStringArgs(args)

		log.BuildStep(name, strings.Join(strArgs, ", "))

		return fn(rk, args, m, self)
	}

	rk.mrb.TopSelf().SingletonClass().DefineMethod(name, hookFunc, args)
}

// Run the script.
func (rk *RubyKube) Run(script string) (*mruby.MrbValue, error) {
	var value *mruby.MrbValue

	value, err := rk.mrb.LoadString(script)
	if err != nil {
		return nil, err
	}

	if _, err := value.Call("inspect"); err != nil {
		return value, fmt.Errorf("could not call `#inspect` [%q]", err)
	}

	getLastValue := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) { return value, nil }

	if value.Type() != mruby.TypeNil {
		rk.mrb.TopSelf().SingletonClass().DefineMethod("_", getLastValue, mruby.ArgsReq(0))
	}

	rk.mrb.TopSelf().SingletonClass().DefineMethod("$?", getLastValue, mruby.ArgsReq(0))

	return value, nil
}

func (rk *RubyKube) SetNamespace(ns string) {
	if ns == "" {
		ns = "*"
	}
	rk.state.Namespace = ns
	rk.readline.SetPrompt(fmt.Sprintf("kubeplay (namespace=%q)> ", rk.state.Namespace))
}

func (rk *RubyKube) GetNamespace() string {
	if rk.state.Namespace == "*" {
		return ""
	}
	return rk.state.Namespace
}

// Close tears down all functions of the RubyKube, preparing it for exit.
func (rk *RubyKube) Close() error {
	rk.mrb.EnableGC()
	rk.mrb.FullGC()
	rk.mrb.Close()
	return nil
}
