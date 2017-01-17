package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type replicaSetClass struct {
	class   *mruby.Class
	objects []replicaSetClassInstance
	rk      *RubyKube
}

type replicaSetClassInstance struct {
	self *mruby.MrbValue
	vars *replicaSetClassInstanceVars
}

type replicaSetClassInstanceVars struct {
	replicaSet replicaSetTypeAlias
}

func newReplicaSetClass(rk *RubyKube) *replicaSetClass {
	c := &replicaSetClass{objects: []replicaSetClassInstance{}, rk: rk}
	c.class = defineReplicaSetClass(rk, c)
	return c
}

func defineReplicaSetClass(rk *RubyKube, c *replicaSetClass) *mruby.Class {
	// common methods
	return rk.defineClass("ReplicaSet", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.replicaSet); err != nil {
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
				return marshalToJSON(vars.replicaSet, m)
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

func (c *replicaSetClass) New() (*replicaSetClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := replicaSetClassInstance{
		self: s,
		vars: &replicaSetClassInstanceVars{
			replicaSetTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *replicaSetClass) LookupVars(this *mruby.MrbValue) (*replicaSetClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "ReplicaSet")
}
