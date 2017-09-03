package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type daemonSetTypeAlias kext.DaemonSet

//go:generate gotemplate "./templates/resource" "daemonSetClass(\"DaemonSet\", daemonSet, daemonSetTypeAlias)"

func (c *daemonSetClass) getSingleton(ns, name string) (*kext.DaemonSet, error) {
	return c.rk.clientset.Extensions().DaemonSets(ns).Get(name)
}

//go:generate gotemplate "./templates/resource/singleton" "daemonSetSingletonModule(daemonSetClass, \"daemonSet\", daemonSet, daemonSetTypeAlias)"

//go:generate gotemplate "./templates/resource/podfinder" "daemonSetPodFinderModule(daemonSetClass, \"daemonSet\", daemonSet, daemonSetTypeAlias)"

func (c *daemonSetClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.definePodFinderMethods()
}

func (o *daemonSetClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
