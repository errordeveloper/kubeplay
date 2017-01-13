package rubykube

/*
  funcs.go provides simple functions for use within rk that do *not*
  They are intended to be used as gathering functions for predicates and templating.
*/

import (
	_ "fmt"
	"io/ioutil"
	"os"
	_ "strings"

	mruby "github.com/mitchellh/go-mruby"
)

type funcDefinition struct {
	fun     funcFunc
	argSpec mruby.ArgSpec
}

type funcFunc func(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

// mrubyJumpTable is the dispatch instructions sent to the mruby interpreter at rk setup.
var funcJumpTable = map[string]funcDefinition{
	"load":   {loadFunc, mruby.ArgsReq(1)},
	"getenv": {getenv, mruby.ArgsReq(1)},
}

// loadFunc implements the load function.
//
// load loads a new ruby file at the point of the function call. it is
// principally used to extend and consolidate reusable code.
func loadFunc(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	args := m.GetArgs()
	if err := checkArgs(args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	content, err := ioutil.ReadFile(args[0].String())
	if err != nil {
		return nil, createException(m, err.Error())
	}

	val, err := rk.Run(string(content))
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return val, nil
}

// getenv retrieves a value from the environment (passed in as string)
// and returns a string with the value. If no value exists, an empty string is
// returned.
func getenv(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	args := m.GetArgs()

	if err := standardCheck(rk, args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	return mruby.String(os.Getenv(args[0].String())), nil
}

const (
	classMethod = iota
	instanceMethod
)

type methodDefintion struct {
	argSpec    mruby.ArgSpec
	methodFunc mruby.Func
	methodType int
}

func (rk *RubyKube) defineClass(name string, methods map[string]methodDefintion) *mruby.Class {
	class := rk.mrb.DefineClass(name, rk.classes.Root)
	rk.appendMethods(class, methods)
	return class
}

func (rk *RubyKube) appendMethods(class *mruby.Class, methods map[string]methodDefintion) {
	for name, m := range methods {
		if m.methodType == classMethod {
			class.DefineClassMethod(name, m.methodFunc, m.argSpec)
		} else {
			class.DefineMethod(name, m.methodFunc, m.argSpec)
		}
	}
}

func defineClass(rk *RubyKube, name string, methods map[string]methodDefintion) *mruby.Class {
	return rk.defineClass(name, methods)
}
