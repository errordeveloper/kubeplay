package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type daemonSetListTypeAlias kext.DaemonSetList

//go:generate gotemplate "./templates/resource" "daemonSetsClass(\"DaemonSets\", daemonSets, daemonSetListTypeAlias)"

func (c *daemonSetsClass) getList(ns string, listOptions meta.ListOptions) (*kext.DaemonSetList, error) {
	return c.rk.clientset.Extensions().DaemonSets(ns).List(listOptions)
}

func (c *daemonSetsClass) getItem(daemonSets daemonSetListTypeAlias, index int) (*daemonSetClassInstance, error) {
	newDaemonSetObj, err := c.rk.classes.DaemonSet.New()
	if err != nil {
		return nil, err
	}
	daemonSet := daemonSets.Items[index]
	newDaemonSetObj.vars.daemonSet = daemonSetTypeAlias(daemonSet)
	return newDaemonSetObj, nil
}

//go:generate gotemplate "./templates/resource/list" "daemonSetsListModule(daemonSetsClass, \"DaemonSets\", daemonSets, daemonSetListTypeAlias)"

func (c *daemonSetsClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *daemonSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
