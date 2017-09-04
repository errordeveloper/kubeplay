package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type deploymentsClass struct {
	class   *mruby.Class
	objects []deploymentsClassInstance
	rk      *RubyKube
}

type deploymentsClassInstance struct {
	self *mruby.MrbValue
	vars *deploymentsClassInstanceVars
}

type deploymentsClassInstanceVars struct {
	deployments deploymentListTypeAlias
}

func newDeploymentsClass(rk *RubyKube) *deploymentsClass {
	c := &deploymentsClass{objects: []deploymentsClassInstance{}, rk: rk}
	c.class = defineDeploymentsClass(rk, c)
	return c
}

func defineDeploymentsClass(rk *RubyKube, c *deploymentsClass) *mruby.Class {
	// common methods
	return rk.defineClass("Deployments", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.deployments); err != nil {
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
				return marshalToJSON(vars.deployments, m)
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

func (c *deploymentsClass) New() (*deploymentsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := deploymentsClassInstance{
		self: s,
		vars: &deploymentsClassInstanceVars{
			deploymentListTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *deploymentsClass) LookupVars(this *mruby.MrbValue) (*deploymentsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Deployments")
}
