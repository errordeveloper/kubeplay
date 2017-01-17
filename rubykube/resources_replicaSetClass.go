package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type replicaSetTypeAlias kext.ReplicaSet

//go:generate gotemplate "./templates/resource" "replicaSetClass(\"ReplicaSet\", replicaSet, replicaSetTypeAlias)"

func (c *replicaSetClass) getSignleton(ns, name string) (*kext.ReplicaSet, error) {
	return c.rk.clientset.Extensions().ReplicaSets(ns).Get(name)
}

//go:generate gotemplate "./templates/resource/singleton" "replicaSetSingletonModule(replicaSetClass, \"replicaSet\", replicaSet, replicaSetTypeAlias)"

func (c *replicaSetClass) defineOwnMethods() {
	c.defineSingletonMethods()
}

func (o *replicaSetClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
