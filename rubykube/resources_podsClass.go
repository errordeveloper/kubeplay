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

func (c *podsClass) getItem(pods podListTypeAlias, index int) (*podClassInstance, error) {
	newPodObj, err := c.rk.classes.Pod.New()
	if err != nil {
		return nil, err
	}
	pod := pods.Items[index]
	newPodObj.vars.pod = podTypeAlias(pod)
	return newPodObj, nil
}

//go:generate gotemplate "./templates/resource/list" "podsListModule(podsClass, \"Pods\", pods, podListTypeAlias)"

func (c *podsClass) defineOwnMethods() {
	c.defineListMethods()
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
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
				i := n.Fixnum()

				if i >= l {
					return nil, nil
				}

				if -(l-1) <= i && i < 0 {
					i = l + i
				} else {
					return nil, nil
				}

				obj, err := c.getItem(vars.pods, i)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return obj.self, nil
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
					obj, err := c.getItem(vars.pods, 0)
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
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
					obj, err := c.getItem(vars.pods, rand.Intn(l))
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
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
					obj, err := c.getItem(vars.pods, l-1)
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
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
