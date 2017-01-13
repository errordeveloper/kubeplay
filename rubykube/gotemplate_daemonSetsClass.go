package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type daemonSetsClass struct {
	class   *mruby.Class
	objects []daemonSetsClassInstance
	rk      *RubyKube
}

type daemonSetsClassInstance struct {
	self *mruby.MrbValue
	vars *daemonSetsClassInstanceVars
}

type daemonSetsClassInstanceVars struct {
	daemonSets *daemonSetList
}

func newDaemonSetsClass(rk *RubyKube) *daemonSetsClass {
	c := &daemonSetsClass{objects: []daemonSetsClassInstance{}, rk: rk}
	c.class = defineDaemonSetsClass(rk, c)
	return c
}

func defineDaemonSetsClass(rk *RubyKube, c *daemonSetsClass) *mruby.Class {
	// common methods
	return rk.defineClass("DaemonSets", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.daemonSets); err != nil {
					return nil, createException(m, err.Error())
				}

				return rbconv.Value(), nil
			},
			instanceMethod,
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return marshalToJSON(vars.daemonSets, m)
			},
			instanceMethod,
		},
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *daemonSetsClass) New() (*daemonSetsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := daemonSetsClassInstance{
		self: s,
		vars: &daemonSetsClassInstanceVars{
			&daemonSetList{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *daemonSetsClass) LookupVars(this *mruby.MrbValue) (*daemonSetsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "DaemonSets")
}
