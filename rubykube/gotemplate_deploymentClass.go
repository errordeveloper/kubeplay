package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type deploymentClass struct {
	class   *mruby.Class
	objects []deploymentClassInstance
	rk      *RubyKube
}

type deploymentClassInstance struct {
	self *mruby.MrbValue
	vars *deploymentClassInstanceVars
}

type deploymentClassInstanceVars struct {
	deployment deploymentTypeAlias
}

func newDeploymentClass(rk *RubyKube) *deploymentClass {
	c := &deploymentClass{objects: []deploymentClassInstance{}, rk: rk}
	c.class = defineDeploymentClass(rk, c)
	return c
}

func defineDeploymentClass(rk *RubyKube, c *deploymentClass) *mruby.Class {
	// common methods
	return rk.defineClass("Deployment", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.deployment); err != nil {
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
				return marshalToJSON(vars.deployment, m)
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

func (c *deploymentClass) New() (*deploymentClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := deploymentClassInstance{
		self: s,
		vars: &deploymentClassInstanceVars{
			deploymentTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *deploymentClass) LookupVars(this *mruby.MrbValue) (*deploymentClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "Deployment")
}
