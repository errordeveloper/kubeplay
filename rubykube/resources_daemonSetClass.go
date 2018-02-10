package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type daemonSetTypeAlias = appsv1.DaemonSet

//go:generate gotemplate "./templates/resource" "daemonSetClass(\"DaemonSet\", daemonSet, daemonSetTypeAlias)"

func (c *daemonSetClass) getSingleton(ns, name string) (*appsv1.DaemonSet, error) {
	return c.rk.clientset.Apps().DaemonSets(ns).Get(name, metav1.GetOptions{})
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
