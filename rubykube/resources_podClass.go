package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podTypeAlias kapi.Pod

//go:generate gotemplate "./templates/resource" "podClass(\"Pod\", pod, podTypeAlias)"

func (c *podClass) getSignleton(ns, name string) (*kapi.Pod, error) {
	return c.rk.clientset.Core().Pods(ns).Get(name)
}

//go:generate gotemplate "./templates/resource/singleton" "podSingletonModule(podClass, \"Pod\", pod, podTypeAlias)"

func (c *podClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"create!": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				pod := kapi.Pod(vars.pod)
				ns := c.rk.GetDefaultNamespace(pod.ObjectMeta.Namespace)

				if _, err = c.rk.clientset.Core().Pods(ns).Create(&pod); err != nil {
					return nil, createException(m, err.Error())
				}

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

				ns := c.rk.GetDefaultNamespace(vars.pod.ObjectMeta.Namespace)

				if err = c.rk.clientset.Core().Pods(ns).Delete(vars.pod.ObjectMeta.Name, &kapi.DeleteOptions{}); err != nil {
					return nil, createException(m, err.Error())
				}

				return self, nil
			},
			instanceMethod,
		},
		"logs": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				body, err := c.rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).GetLogs(vars.pod.ObjectMeta.Name, &kapi.PodLogOptions{}).Do().Raw()
				if err != nil {
					return nil, createException(m, err.Error())
				}

				fmt.Println(string(body))

				return nil, nil

			},
			instanceMethod,
		},
	})
}

func (o *podClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
