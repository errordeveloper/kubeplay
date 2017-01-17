package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type daemonSetClass struct {
	class   *mruby.Class
	objects []daemonSetClassInstance
	rk      *RubyKube
}

type daemonSetClassInstance struct {
	self *mruby.MrbValue
	vars *daemonSetClassInstanceVars
}

type daemonSetClassInstanceVars struct {
	daemonSet daemonSetTypeAlias
}

func newDaemonSetClass(rk *RubyKube) *daemonSetClass {
	c := &daemonSetClass{objects: []daemonSetClassInstance{}, rk: rk}
	c.class = defineDaemonSetClass(rk, c)
	return c
}

func defineDaemonSetClass(rk *RubyKube, c *daemonSetClass) *mruby.Class {
	// common methods
	return rk.defineClass("DaemonSet", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.daemonSet); err != nil {
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
				return marshalToJSON(vars.daemonSet, m)
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

func (c *daemonSetClass) New() (*daemonSetClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := daemonSetClassInstance{
		self: s,
		vars: &daemonSetClassInstanceVars{
			daemonSetTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *daemonSetClass) LookupVars(this *mruby.MrbValue) (*daemonSetClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "DaemonSet")
}
