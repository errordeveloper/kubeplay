package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type replicaSetListTypeAlias kext.ReplicaSetList

//go:generate gotemplate "./templates/resource" "replicaSetsClass(\"ReplicaSets\", replicaSets, replicaSetListTypeAlias)"

func (c *replicaSetsClass) getList(ns string, listOptions kapi.ListOptions) (*kext.ReplicaSetList, error) {
	return c.rk.clientset.Extensions().ReplicaSets(ns).List(listOptions)
}

//go:generate gotemplate "./templates/resource/list" "replicaSetsListModule(replicaSetsClass, \"ReplicaSets\", replicaSets, replicaSetListTypeAlias)"

func (c *replicaSetsClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *replicaSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
