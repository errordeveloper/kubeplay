package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podTypeAlias kapi.Pod

//go:generate gotemplate "./templates/resource" "podClass(\"Pod\", pod, podTypeAlias)"

func (c *podClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				pod, err := c.rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).Get(vars.pod.ObjectMeta.Name)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				p := podTypeAlias(*pod)
				vars.pod = &p
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

				fmt.Printf("self: %s/%s\n", vars.pod.ObjectMeta.Namespace, vars.pod.ObjectMeta.Name)
				return self, nil
			},
			instanceMethod,
		},
		"delete!": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if err = c.rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).Delete(vars.pod.ObjectMeta.Name, &kapi.DeleteOptions{}); err != nil {
					return nil, createException(m, err.Error())
				}

				return self, nil
			},
			instanceMethod,
		},
	})
}

func (o *podClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
