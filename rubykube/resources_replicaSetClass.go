package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type replicaSetTypeAlias kext.ReplicaSet

//go:generate gotemplate "./templates/resource" "replicaSetClass(\"ReplicaSet\", replicaSet, replicaSetTypeAlias)"

func (c *replicaSetClass) getSingleton(ns, name string) (*kext.ReplicaSet, error) {
	return c.rk.clientset.Extensions().ReplicaSets(ns).Get(name, meta.GetOptions{})
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
