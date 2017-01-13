package rubykube

import (
	"fmt"
	_ "math/rand"

	mruby "github.com/mitchellh/go-mruby"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type daemonSetListTypeAlias kext.DaemonSetList

//go:generate gotemplate "./templates/resource" "daemonSetsClass(\"DaemonSets\", daemonSets, daemonSetListTypeAlias)"

func (c *daemonSetsClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, daemonSetNameRegexp, listOptions, err := c.rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				daemonSets, err := c.rk.clientset.Extensions().DaemonSets(c.rk.GetNamespace(ns)).List(*listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if daemonSetNameRegexp != nil {
					for _, daemonSet := range daemonSets.Items {
						if daemonSetNameRegexp.MatchString(daemonSet.ObjectMeta.Name) {
							vars.daemonSets.Items = append(vars.daemonSets.Items, daemonSet)
						}
					}
				} else {
					for _, c := range daemonSets.Items {
						vars.daemonSets.Items = append(vars.daemonSets.Items, c)
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

				for n, daemonSet := range vars.daemonSets.Items {
					fmt.Printf("%d: %s/%s\n", n, daemonSet.ObjectMeta.Namespace, daemonSet.ObjectMeta.Name)
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

				return m.FixnumValue(len(vars.daemonSets.Items)), nil
			},
			instanceMethod,
		},
	})
}

func (o *daemonSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
