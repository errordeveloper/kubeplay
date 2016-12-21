package rubykube

/*
  verbs.go is a collection of the verbs used to do stuff.
*/

import (
	_ "fmt"

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
	"new":  {newMaker, mruby.ArgsReq(1)},
	"pods": {pods, mruby.ArgsReq(0)},
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

	if value, err = newPodsObj.Update(); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}
