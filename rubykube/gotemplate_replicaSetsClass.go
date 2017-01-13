package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, instanceVariableName, instanceVariableType)

type replicaSetsClass struct {
	class   *mruby.Class
	objects []replicaSetsClassInstance
	rk      *RubyKube
}

type replicaSetsClassInstance struct {
	self *mruby.MrbValue
	vars *replicaSetsClassInstanceVars
}

type replicaSetsClassInstanceVars struct {
	replicaSets replicaSetListTypeAlias
}

func newReplicaSetsClass(rk *RubyKube) *replicaSetsClass {
	c := &replicaSetsClass{objects: []replicaSetsClassInstance{}, rk: rk}
	c.class = defineReplicaSetsClass(rk, c)
	return c
}

func defineReplicaSetsClass(rk *RubyKube, c *replicaSetsClass) *mruby.Class {
	// common methods
	return rk.defineClass("ReplicaSets", map[string]methodDefintion{
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.replicaSets); err != nil {
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
				return marshalToJSON(vars.replicaSets, m)
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

func (c *replicaSetsClass) New() (*replicaSetsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := replicaSetsClassInstance{
		self: s,
		vars: &replicaSetsClassInstanceVars{
			replicaSetListTypeAlias{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *replicaSetsClass) LookupVars(this *mruby.MrbValue) (*replicaSetsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "ReplicaSets")
}
