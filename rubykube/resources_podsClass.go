package rubykube

import (
	"math/rand"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podListTypeAlias kapi.PodList

//go:generate gotemplate "./templates/resource" "podsClass(\"Pods\", pods, podListTypeAlias)"

func (c *podsClass) getList(ns string, listOptions kapi.ListOptions) (*kapi.PodList, error) {
	return c.rk.clientset.Core().Pods(ns).List(listOptions)
}

//go:generate gotemplate "./templates/resource/list" "podsListModule(podsClass, \"Pods\", pods, podListTypeAlias)"

func (c *podsClass) defineOwnMethods() {
	c.defineListMethods()
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				var pod kapi.Pod
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()
				err = standardCheck(c.rk, args, 1)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				n := args[0]
				if n.Type() != mruby.TypeFixnum {
					return nil, createException(m, "Argument must be an integer")
				}

				l := len(vars.pods.Items)

				if n.Fixnum() >= l {
					return nil, nil
				}

				if n.Fixnum() >= 0 {
					pod = vars.pods.Items[n.Fixnum()]
				} else if -(l-1) <= n.Fixnum() && n.Fixnum() < 0 {
					pod = vars.pods.Items[l+n.Fixnum()]
				} else {
					return nil, nil
				}

				newPodObj, err := c.rk.classes.Pod.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				newPodObj.vars.pod = podTypeAlias(pod)
				return newPodObj.self, nil
			},
			instanceMethod,
		},
		"first": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if len(vars.pods.Items) > 0 {
					newPodObj, err := c.rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = podTypeAlias(vars.pods.Items[0])
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"any": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.pods.Items)
				if l > 0 {
					newPodObj, err := c.rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = podTypeAlias(vars.pods.Items[rand.Intn(l)])
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"last": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.pods.Items)

				if l > 0 {
					newPodObj, err := c.rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = podTypeAlias(vars.pods.Items[l-1])
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
	})
}

func (o *podsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
