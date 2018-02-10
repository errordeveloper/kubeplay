package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type daemonSetPodFinderModule struct{}

func (c *daemonSetClass) definePodFinderMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"pods": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns := vars.daemonSet.ObjectMeta.Namespace

				selector := []string{}
				// TODO: probably should use `spec.selector`
				for k, v := range vars.daemonSet.ObjectMeta.Labels {
					selector = append(selector, fmt.Sprintf("%s in (%s)", k, v))
				}
				listOptions := metav1.ListOptions{LabelSelector: strings.Join(selector, ",")}

				pods, err := c.rk.clientset.Core().Pods(ns).List(listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				// TODO: verify `ownerReferences`...

				newPodsObj, err := c.rk.classes.Pods.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				newPodsObj.vars.pods = podListTypeAlias(*pods)
				return newPodsObj.self, nil
			},
			instanceMethod,
		},
	})
}
