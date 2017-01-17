package rubykube

import (
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
}

func (o *podsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
