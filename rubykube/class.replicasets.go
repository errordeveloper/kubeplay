package rubykube

import (
	"fmt"
	_ "math/rand"

	mruby "github.com/mitchellh/go-mruby"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type replicaSetListTypeAlias kext.ReplicaSetList

//go:generate gotemplate "./templates/resource" "replicaSetsClass(\"ReplicaSets\", replicaSets, replicaSetListTypeAlias)"

func (c *replicaSetsClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, replicaSetNameRegexp, listOptions, err := c.rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				replicaSets, err := c.rk.clientset.Extensions().ReplicaSets(c.rk.GetNamespace(ns)).List(*listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if replicaSetNameRegexp != nil {
					for _, replicaSet := range replicaSets.Items {
						if replicaSetNameRegexp.MatchString(replicaSet.ObjectMeta.Name) {
							vars.replicaSets.Items = append(vars.replicaSets.Items, replicaSet)
						}
					}
				} else {
					for _, replicaSet := range replicaSets.Items {
						vars.replicaSets.Items = append(vars.replicaSets.Items, replicaSet)
					}
				}
				return self, nil
			},
			instanceMethod,
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
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
		"count": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.FixnumValue(len(vars.replicaSets.Items)), nil
			},
			instanceMethod,
		},
	})
}

func (o *replicaSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
