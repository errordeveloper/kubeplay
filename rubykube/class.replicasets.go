package rubykube

import (
	"fmt"
	_ "math/rand"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
	//kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

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
	replicaSets *kext.ReplicaSetList
}

func newReplicaSetsClass(rk *RubyKube) *replicaSetsClass {
	c := &replicaSetsClass{objects: []replicaSetsClassInstance{}, rk: rk}
	c.class = defineReplicaSetsClass(rk, c)
	return c
}

func defineReplicaSetsClass(rk *RubyKube, r *replicaSetsClass) *mruby.Class {
	return defineClass(rk, "RepicaSets", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, replicaSetNameRegexp, listOptions, err := rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				replicaSets, err := rk.clientset.Extensions().ReplicaSets(rk.GetNamespace(ns)).List(*listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if replicaSetNameRegexp != nil {
					for _, r := range replicaSets.Items {
						if replicaSetNameRegexp.MatchString(r.ObjectMeta.Name) {
							vars.replicaSets.Items = append(vars.replicaSets.Items, r)
						}
					}
				} else {
					vars.replicaSets = replicaSets
				}
				return self, nil
			},
			instanceMethod,
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				for n, replicaSet := range vars.replicaSets.Items {
					fmt.Printf("%d: %s/%s\n", n, replicaSet.ObjectMeta.Namespace, replicaSet.ObjectMeta.Name)
				}
				return self, nil
			},
			instanceMethod,
		},
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
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
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return marshalToJSON(vars.replicaSets, m)
			},
			instanceMethod,
		},
		"count": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.FixnumValue(len(vars.replicaSets.Items)), nil
			},
			instanceMethod,
		},
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(r.objects)), nil
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
			&kext.ReplicaSetList{},
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
	return nil, fmt.Errorf("could not find class instance")
}

func (o *replicaSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
