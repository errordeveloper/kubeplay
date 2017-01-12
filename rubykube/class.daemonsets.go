package rubykube

import (
	"fmt"
	_ "math/rand"

	"github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
	//kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type daemonSetsClass struct {
	class   *mruby.Class
	objects []daemonSetsClassInstance
	rk      *RubyKube
}

type daemonSetsClassInstance struct {
	self *mruby.MrbValue
	vars *daemonSetsClassInstanceVars
}

type daemonSetsClassInstanceVars struct {
	daemonSets *kext.DaemonSetList
}

func newDaemonSetsClass(rk *RubyKube) *daemonSetsClass {
	c := &daemonSetsClass{objects: []daemonSetsClassInstance{}, rk: rk}
	c.class = defineDaemonSetsClass(rk, c)
	return c
}

func defineDaemonSetsClass(rk *RubyKube, r *daemonSetsClass) *mruby.Class {
	return defineClass(rk, "RepicaSets", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, daemonSetNameRegexp, listOptions, err := rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				daemonSets, err := rk.clientset.Extensions().DaemonSets(rk.GetNamespace(ns)).List(*listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if daemonSetNameRegexp != nil {
					for _, r := range daemonSets.Items {
						if daemonSetNameRegexp.MatchString(r.ObjectMeta.Name) {
							vars.daemonSets.Items = append(vars.daemonSets.Items, r)
						}
					}
				} else {
					vars.daemonSets = daemonSets
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

				for n, daemonSet := range vars.daemonSets.Items {
					fmt.Printf("%d: %s/%s\n", n, daemonSet.ObjectMeta.Namespace, daemonSet.ObjectMeta.Name)
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
				if err := rbconv.Convert(vars.daemonSets); err != nil {
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
				return marshalToJSON(vars.daemonSets, m)
			},
			instanceMethod,
		},
		"count": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := r.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.FixnumValue(len(vars.daemonSets.Items)), nil
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

func (c *daemonSetsClass) New() (*daemonSetsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := daemonSetsClassInstance{
		self: s,
		vars: &daemonSetsClassInstanceVars{
			&kext.DaemonSetList{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *daemonSetsClass) LookupVars(this *mruby.MrbValue) (*daemonSetsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *daemonSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
