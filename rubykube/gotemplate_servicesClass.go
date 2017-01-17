package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type servicesClass struct {
	class   *mruby.Class
	objects []servicesClassInstance
	rk      *RubyKube
}

type servicesClassInstance struct {
	self *mruby.MrbValue
	vars *servicesClassInstanceVars
}

type servicesClassInstanceVars struct {
	services serviceListTypeAlias
}

func newServicesClass(rk *RubyKube) *servicesClass {
	c := &servicesClass{objects: []servicesClassInstance{}, rk: rk}
	c.class = defineServicesClass(rk, c)
	return c
}

func defineServicesClass(rk *RubyKube, c *servicesClass) *mruby.Class {
	// common methods
	return rk.defineClass("Services", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.services); err != nil {
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
				return marshalToJSON(vars.services, m)
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

func (c *servicesClass) New() (*servicesClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := servicesClassInstance{
		self: s,
		vars: &servicesClassInstanceVars{
			serviceListTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *servicesClass) LookupVars(this *mruby.MrbValue) (*servicesClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Services")
}
