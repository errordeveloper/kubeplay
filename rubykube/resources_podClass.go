package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

type podTypeAlias = corev1.Pod

//go:generate gotemplate "./templates/resource" "podClass(\"Pod\", pod, podTypeAlias)"

func (c *podClass) getSingleton(ns, name string) (*corev1.Pod, error) {
	return c.rk.clientset.Core().Pods(ns).Get(name, metav1.GetOptions{})
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

				pod := corev1.Pod(vars.pod)
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

				if err = c.rk.clientset.Core().Pods(ns).Delete(vars.pod.ObjectMeta.Name, &metav1.DeleteOptions{}); err != nil {
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

				newPodLogsObj, err := c.rk.classes.PodLogs.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				pod := corev1.Pod(vars.pod)
				newPodLogsObj.vars.pods = []corev1.Pod{pod}
				return callWithException(m, newPodLogsObj.self, "get!")
			},
			instanceMethod,
		},
	})
}

func (o *podClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
