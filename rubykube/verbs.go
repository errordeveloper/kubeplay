package rubykube

/*
  verbs.go is a collection of the verbs used to do stuff.
*/

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// Definition is a jump table definition used for programming the DSL into the
// mruby interpreter.
type verbDefinition struct {
	verbFunc verbFunc
	argSpec  mruby.ArgSpec
}

// verbJumpTable is the dispatch instructions sent to the builder at preparation time.
var verbJumpTable map[string]verbDefinition

func init() {
	// This is to avoid initialisation loop error from the compiler
	verbJumpTable = map[string]verbDefinition{
		"pods":                {pods, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"services":            {services, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"replicasets":         {replicaSets, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"daemonsets":          {daemonSets, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"make_pod":            {makePod, mruby.ArgsReq(1)},
		"make_label_selector": {makeLabelSelector, mruby.ArgsReq(1)},
		"make_field_selector": {makeFieldSelector, mruby.ArgsReq(1)},
		"using":               {using, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"namespace":           {namespace, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
		"def_alias":           {defAlias, mruby.ArgsReq(2)},
	}
}

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newPodsObj, err := rk.classes.Pods.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newPodsObj.Update(args...); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}

func services(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newServicesObj, err := rk.classes.Services.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newServicesObj.Update(args...); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}

func replicaSets(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newReplicaSetsObj, err := rk.classes.ReplicaSets.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newReplicaSetsObj.Update(args...); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}

func daemonSets(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newReplicaSetsObj, err := rk.classes.DaemonSets.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newReplicaSetsObj.Update(args...); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}

func makePod(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		err error
	)

	newPodMakerObj, err := rk.classes.PodMaker.New()

	if err != nil {
		return nil, createException(m, err.Error())
	}

	value, err := newPodMakerObj.self.Call("pod!", toValues(args)...)
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return value, nil
}

func makeLabelSelector(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if err := checkArgs(args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	newLabelNameObj, err := rk.classes.LabelSelector.New(toValues(args)...)
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return newLabelNameObj.self, nil
}

func makeFieldSelector(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if err := checkArgs(args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	newFieldNameObj, err := rk.classes.FieldSelector.New(toValues(args)...)
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return newFieldNameObj.self, nil
}

func using(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if len(args) == 0 {
		fmt.Println("%+v", rk.state)
		return nil, nil

	}

	if args[0].Type() != mruby.TypeHash {
		return nil, createException(m, "First argument must be a hash")
	}

	pc, err := NewParamsCollection(args[0],
		params{
			allowed:   []string{"namespace", "cluster", "context"},
			required:  []string{"namespace"},
			valueType: mruby.TypeString,
		},
	)

	if err != nil {
		return nil, createException(m, err.Error())
	}

	p := pc.ToMapOfStrings()

	rk.SetNamespace(p["namespace"])

	if len(args) == 2 {
		//TODO: allow calling with a block
	}

	return nil, nil
}

func namespace(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if len(args) == 0 {
		return m.StringValue(rk.state.Namespace), nil
	}

	if args[0].Type() != mruby.TypeString {
		return nil, createException(m, "First argument must be a string")
	}

	ns := ValidString(args[0])
	if ns == nil {
		return nil, createException(m, "First argument must be a non-empty string")
	}

	rk.SetNamespace(*ns)

	if len(args) == 2 {
		//TODO: allow calling with a block
	}

	return nil, nil
}

func defAlias(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if len(args) != 2 {
		return nil, createException(m, "Must provide exactly two arguments")
	}

	if args[0].Type() != mruby.TypeSymbol && args[1].Type() != mruby.TypeSymbol {
		return nil, createException(m, "Both arguments must be symbols")
	}

	alias := args[0].String()
	verb := args[1].String()

	if _, ok := verbJumpTable[verb]; !ok {
		return nil, createException(m, fmt.Sprintf("Unknown verb %q", verb))
	}

	aliasFunc := func(m *mruby.Mrb, _ *mruby.MrbValue) (mruby.Value, mruby.Value) {
		value, err := self.Call(verb, toValues(m.GetArgs())...)
		if err != nil {
			return nil, createException(m, err.Error())
		}
		return value, nil
	}

	rk.mrb.TopSelf().SingletonClass().DefineMethod(alias, aliasFunc, mruby.ArgsAny())

	return nil, nil
}
