package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type serviceClass struct {
	class   *mruby.Class
	objects []serviceClassInstance
	rk      *RubyKube
}

type serviceClassInstance struct {
	self *mruby.MrbValue
	vars *serviceClassInstanceVars
}

type serviceClassInstanceVars struct {
	service serviceTypeAlias
}

func newServiceClass(rk *RubyKube) *serviceClass {
	c := &serviceClass{objects: []serviceClassInstance{}, rk: rk}
	c.class = defineServiceClass(rk, c)
	return c
}

func defineServiceClass(rk *RubyKube, c *serviceClass) *mruby.Class {
	// common methods
	return rk.defineClass("Service", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.service); err != nil {
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
				return marshalToJSON(vars.service, m)
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

func (c *serviceClass) New() (*serviceClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := serviceClassInstance{
		self: s,
		vars: &serviceClassInstanceVars{
			serviceTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *serviceClass) LookupVars(this *mruby.MrbValue) (*serviceClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Service")
}
