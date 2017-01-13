package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type podsClass struct {
	class   *mruby.Class
	objects []podsClassInstance
	rk      *RubyKube
}

type podsClassInstance struct {
	self *mruby.MrbValue
	vars *podsClassInstanceVars
}

type podsClassInstanceVars struct {
	pods *podListTypeAlias
}

func newPodsClass(rk *RubyKube) *podsClass {
	c := &podsClass{objects: []podsClassInstance{}, rk: rk}
	c.class = definePodsClass(rk, c)
	return c
}

func definePodsClass(rk *RubyKube, c *podsClass) *mruby.Class {
	// common methods
	return rk.defineClass("Pods", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.pods); err != nil {
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
				return marshalToJSON(vars.pods, m)
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

func (c *podsClass) New() (*podsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := podsClassInstance{
		self: s,
		vars: &podsClassInstanceVars{
			&podListTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podsClass) LookupVars(this *mruby.MrbValue) (*podsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Pods")
}
