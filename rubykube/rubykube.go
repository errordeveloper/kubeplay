package rubykube

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/chzyer/readline"
	//"github.com/erikh/box/signal"
	mruby "github.com/mitchellh/go-mruby"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", os.ExpandEnv("${HOME}/.kube/config"), "absolute path to the kubeconfig file")

type RubyKube struct {
	mrb       *mruby.Mrb
	clientset *kubernetes.Clientset
	classes   Classes
	readline  *readline.Instance
	state     *CurrentState
}

type Classes struct {
	Root *mruby.Class

	Pods *podsClass
	Pod  *podClass

	Service  *serviceClass
	Services *servicesClass

	Deployments *deploymentsClass
	Deployment  *deploymentClass

	ReplicaSets *replicaSetsClass
	ReplicaSet  *replicaSetClass

	DaemonSets *daemonSetsClass
	DaemonSet  *daemonSetClass

	PodLogs *podLogsClass

	PodMaker *podMakerClass

	LabelSelector  *labelSelectorClass
	LabelCollector *labelCollectorClass
	LabelKey       *labelKeyClass

	FieldSelector  *fieldSelectorClass
	FieldCollector *fieldCollectorClass
	FieldKey       *fieldKeyClass
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

	fmt.Printf("kubeconfig=%+v\n", config)

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

	//signal.SignalHandler(nil)

	rk.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("kubernetes.NewForConfig: %v", err))
	}

	rk.classes = Classes{Root: rk.mrb.DefineClass("RubyKube", nil)}

	rk.classes.Pods = newPodsClass(rk)
	rk.classes.Pods.defineOwnMethods()

	rk.classes.Pod = newPodClass(rk)
	rk.classes.Pod.defineOwnMethods()

	rk.classes.Services = newServicesClass(rk)
	rk.classes.Services.defineOwnMethods()

	rk.classes.Service = newServiceClass(rk)
	rk.classes.Service.defineOwnMethods()

	rk.classes.Deployments = newDeploymentsClass(rk)
	rk.classes.Deployments.defineOwnMethods()

	rk.classes.Deployment = newDeploymentClass(rk)
	rk.classes.Deployment.defineOwnMethods()

	rk.classes.ReplicaSets = newReplicaSetsClass(rk)
	rk.classes.ReplicaSets.defineOwnMethods()

	rk.classes.ReplicaSet = newReplicaSetClass(rk)
	rk.classes.ReplicaSet.defineOwnMethods()

	rk.classes.DaemonSets = newDaemonSetsClass(rk)
	rk.classes.DaemonSets.defineOwnMethods()

	rk.classes.DaemonSet = newDaemonSetClass(rk)
	rk.classes.DaemonSet.defineOwnMethods()

	rk.classes.PodLogs = newPodLogsClass(rk)
	rk.classes.PodLogs.defineOwnMethods()

	rk.classes.PodMaker = newPodMakerClass(rk)
	rk.classes.PodMaker.defineOwnMethods()

	rk.classes.LabelSelector = newLabelSelectorClass(rk)
	rk.classes.LabelSelector.defineOwnMethods()

	rk.classes.LabelCollector = newLabelCollectorClass(rk)
	rk.classes.LabelCollector.defineOwnMethods()

	rk.classes.LabelKey = newLabelKeyClass(rk)
	rk.classes.LabelKey.defineOwnMethods()

	rk.classes.FieldSelector = newFieldSelectorClass(rk)
	rk.classes.FieldSelector.defineOwnMethods()

	rk.classes.FieldCollector = newFieldCollectorClass(rk)
	rk.classes.FieldCollector.defineOwnMethods()

	rk.classes.FieldKey = newFieldKeyClass(rk)
	rk.classes.FieldKey.defineOwnMethods()

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

// RunCode runs the ruby value (a proc) and returns the result.
func (rk *RubyKube) RunCode(block *mruby.MrbValue, stackKeep int) (*mruby.MrbValue, int, error) {
	var value *mruby.MrbValue

	keep, value, err := rk.mrb.RunWithContext(block, rk.mrb.TopSelf(), stackKeep)
	if err != nil {
		return nil, keep, err
	}

	if _, err := value.Call("inspect"); err != nil {
		return value, keep, fmt.Errorf("could not call `#inspect` [%q]", err)
	}

	getLastValue := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) { return value, nil }

	if value.Type() != mruby.TypeNil {
		rk.mrb.TopSelf().SingletonClass().DefineMethod("_", getLastValue, mruby.ArgsReq(0))
	}

	rk.mrb.TopSelf().SingletonClass().DefineMethod("$?", getLastValue, mruby.ArgsReq(0))

	return value, keep, nil

}

func (rk *RubyKube) NormalPrompt() {
	rk.readline.SetPrompt(fmt.Sprintf("kubeplay (namespace=%q)> ", rk.state.Namespace))
}

func (rk *RubyKube) MultiLinePrompt() {
	rk.readline.SetPrompt(fmt.Sprintf("kubeplay (namespace=%q)> ....| ", rk.state.Namespace))
}

func (rk *RubyKube) SetNamespace(ns string) {
	if ns == "" {
		ns = "*"
	}
	rk.state.Namespace = ns
	rk.readline.SetPrompt(fmt.Sprintf("kubeplay (namespace=%q)> ", rk.state.Namespace))
}

func (rk *RubyKube) GetNamespace(override string) string {
	ns := rk.state.Namespace
	if override != "" {
		ns = override
	}

	if ns == "*" {
		return ""
	}
	return ns
}

func (rk *RubyKube) GetDefaultNamespace(override string) string {
	ns := rk.state.Namespace
	if override != "" {
		ns = override
	}

	if ns == "*" {
		return "default"
	}
	return ns
}

func (rk *RubyKube) resourceArgs(args []*mruby.MrbValue) (string, *regexp.Regexp, *metav1.ListOptions, error) {
	var (
		ns          string
		nameRegexp  *regexp.Regexp
		listOptions metav1.ListOptions
	)

	validName := "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	namespacePrefix := fmt.Sprintf(`?P<namespace>%s|\*`, validName)

	// pods "foo/"
	// pods "*/"
	allWithinNamespace := regexp.MustCompile(
		fmt.Sprintf(`^(%s)\/(\*)?$`,
			namespacePrefix))
	// pods "*-bar"
	nameBeginsWith := regexp.MustCompile(
		fmt.Sprintf(`^((%s)\/)?(?P<name>%s(-)?)\*$`,
			namespacePrefix,
			validName))
	// pods "bar-*"
	nameEndsWith := regexp.MustCompile(
		fmt.Sprintf(`^((%s)\/)?\*(?P<name>(-)?%s)$`,
			namespacePrefix,
			validName))
	// pods "*-bar-*"
	nameContains := regexp.MustCompile(
		fmt.Sprintf(`^((%s)\/)?\*(?P<name>(-)?%s(-)?)\*$`,
			namespacePrefix,
			validName))

	// TODO pods "*-foo-*-bar-*"

	hasNameGlob := false
	hasSelectors := false

	parseNameGlob := func(arg *mruby.MrbValue) error {
		s := arg.String()
		var p string
		switch {
		case allWithinNamespace.MatchString(s):
			getNamedMatch(allWithinNamespace, s, "namespace", &ns)
		case nameBeginsWith.MatchString(s):
			getNamedMatch(nameBeginsWith, s, "namespace", &ns)
			getNamedMatch(nameBeginsWith, s, "name", &p)
			nameRegexp = regexp.MustCompile(fmt.Sprintf("^(%s)-?(%s)$", p, validName))
		case nameEndsWith.MatchString(s):
			getNamedMatch(nameEndsWith, s, "namespace", &ns)
			getNamedMatch(nameEndsWith, s, "name", &p)
			nameRegexp = regexp.MustCompile(fmt.Sprintf("^(%s)-?(%s)$", validName, p))
		case nameContains.MatchString(s):
			getNamedMatch(nameContains, s, "namespace", &ns)
			getNamedMatch(nameContains, s, "name", &p)
			nameRegexp = regexp.MustCompile(fmt.Sprintf("^(%s)?-?(%s)-?(%s)?$", validName, p, validName))
		default:
			if s != "*" {
				return fmt.Errorf("Invalid glob expression - try `pods \"<namespace>/\"`, `pods \"*/\"` or `pods \"*/foo-*\"`\n")
			}
		}

		hasNameGlob = true
		return nil
	}

	evalLabelSelector := func(block *mruby.MrbValue) error {
		newLabelNameObj, err := rk.classes.LabelSelector.New(block)
		if err != nil {
			return err
		}

		listOptions.LabelSelector = newLabelNameObj.self.String()

		return nil
	}

	evalFieldSelector := func(block *mruby.MrbValue) error {
		newFieldNameObj, err := rk.classes.FieldSelector.New(block)
		if err != nil {
			return err
		}

		listOptions.FieldSelector = newFieldNameObj.self.String()

		return nil
	}

	parseSelectors := func(arg *mruby.MrbValue) error {
		stringCollection, err := NewParamsCollection(arg,
			params{
				allowed:   []string{"labels", "fields"},
				required:  []string{},
				valueType: mruby.TypeString,
				procHandlers: map[string]paramProcHandler{
					"labels": evalLabelSelector,
					"fields": evalFieldSelector,
				},
			},
		)

		if err != nil {
			return err
		}

		p := stringCollection.ToMapOfStrings()

		if v, ok := p["labels"]; ok {
			listOptions.LabelSelector = v
		}
		if v, ok := p["fields"]; ok {
			listOptions.FieldSelector = v
		}

		hasSelectors = true
		return nil
	}

	fail := func(err error) (string, *regexp.Regexp, *metav1.ListOptions, error) {
		return "", nil, nil, err
	}

	secondArgError := func(kind string) error {
		return fmt.Errorf("Found second " + kind + " argument, only single one is allowed - use array notation for mulptiple queries")
	}

	if len(args) >= 1 {
		switch args[0].Type() {
		case mruby.TypeString:
			if err := parseNameGlob(args[0]); err != nil {
				return fail(err)
			}
		case mruby.TypeHash:
			if err := parseSelectors(args[0]); err != nil {
				return fail(err)
			}
		case mruby.TypeProc:
			if err := evalLabelSelector(args[0]); err != nil {
				return fail(err)
			}
			if err := evalFieldSelector(args[0]); err != nil {
				return fail(err)
			}
		case mruby.TypeArray:
			// TODO: we could allow users to collect object matching multiple globs
			return fail(fmt.Errorf("Not yet implemented!"))
		}
	}

	if len(args) >= 2 {

		switch args[1].Type() {
		case mruby.TypeString:
			if hasNameGlob {
				return fail(secondArgError("glob"))

			}
			if err := parseNameGlob(args[1]); err != nil {
				return fail(err)
			}
		case mruby.TypeHash:
			if hasSelectors {
				return fail(secondArgError("selector"))
			}
			if err := parseSelectors(args[1]); err != nil {
				return fail(err)
			}
		case mruby.TypeProc:
			if hasSelectors {
				return fail(secondArgError("selector"))
			}
			if err := evalLabelSelector(args[1]); err != nil {
				return fail(err)
			}
			if err := evalFieldSelector(args[0]); err != nil {
				return fail(err)
			}
		case mruby.TypeArray:
			return fail(fmt.Errorf("Only single array argument is allowed"))
		}
	}

	if len(args) >= 3 {
		return fail(fmt.Errorf("Maximum 2 arguments allowed"))
	}

	return ns, nameRegexp, &listOptions, nil
}

// Mrb returns the mrb (mruby) instance the builder is using.
func (rk *RubyKube) Mrb() *mruby.Mrb {
	return rk.mrb
}

// Close tears down all functions of the RubyKube, preparing it for exit.
func (rk *RubyKube) Close() error {
	rk.mrb.EnableGC()
	rk.mrb.FullGC()
	rk.mrb.Close()
	return nil
}
