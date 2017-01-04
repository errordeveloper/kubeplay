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
var verbJumpTable = map[string]verbDefinition{
	"new":       {newMaker, mruby.ArgsReq(1)},
	"pods":      {pods, mruby.ArgsReq(0)},
	"using":     {using, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
	"namespace": {namespace, mruby.ArgsReq(0) | mruby.ArgsOpt(2)},
}

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

func newMaker(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		err error
	)

	newPodMakerObj, err := rk.classes.PodMaker.New()

	if err != nil {
		return nil, createException(m, err.Error())
	}

	return newPodMakerObj.self, nil
}

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newPodsObj, err := rk.classes.Pods.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	argv := []mruby.Value{}
	for _, arg := range args {
		argv = append(argv, mruby.Value(arg))
	}

	if value, err = newPodsObj.Update(argv...); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
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
