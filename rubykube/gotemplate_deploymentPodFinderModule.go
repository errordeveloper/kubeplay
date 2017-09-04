package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type deploymentPodFinderModule struct{}

func (c *deploymentClass) definePodFinderMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"pods": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns := vars.deployment.ObjectMeta.Namespace

				selector := []string{}
				for k, v := range vars.deployment.ObjectMeta.Labels {
					selector = append(selector, fmt.Sprintf("%s in (%s)", k, v))
				}
				listOptions := kapi.ListOptions{LabelSelector: strings.Join(selector, ",")}

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
