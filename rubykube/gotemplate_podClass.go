package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type podClass struct {
	class   *mruby.Class
	objects []podClassInstance
	rk      *RubyKube
}

type podClassInstance struct {
	self *mruby.MrbValue
	vars *podClassInstanceVars
}

type podClassInstanceVars struct {
	pod podTypeAlias
}

func newPodClass(rk *RubyKube) *podClass {
	c := &podClass{objects: []podClassInstance{}, rk: rk}
	c.class = definePodClass(rk, c)
	return c
}

func definePodClass(rk *RubyKube, c *podClass) *mruby.Class {
	// common methods
	return rk.defineClass("Pod", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.pod); err != nil {
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
				return marshalToJSON(vars.pod, m)
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

func (c *podClass) New() (*podClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := podClassInstance{
		self: s,
		vars: &podClassInstanceVars{
			podTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podClass) LookupVars(this *mruby.MrbValue) (*podClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Pod")
}
