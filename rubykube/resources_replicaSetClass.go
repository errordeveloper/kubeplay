package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type replicaSetTypeAlias appsv1.ReplicaSet

//go:generate gotemplate "./templates/resource" "replicaSetClass(\"ReplicaSet\", replicaSet, replicaSetTypeAlias)"

func (c *replicaSetClass) getSingleton(ns, name string) (*appsv1.ReplicaSet, error) {
	return c.rk.clientset.Apps().ReplicaSets(ns).Get(name, metav1.GetOptions{})
}

//go:generate gotemplate "./templates/resource/singleton" "replicaSetSingletonModule(replicaSetClass, \"replicaSet\", replicaSet, replicaSetTypeAlias)"

//go:generate gotemplate "./templates/resource/podfinder" "replicaSetPodFinderModule(replicaSetClass, \"replicaSet\", replicaSet, replicaSetTypeAlias)"

func (c *replicaSetClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.definePodFinderMethods()
}

func (o *replicaSetClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
