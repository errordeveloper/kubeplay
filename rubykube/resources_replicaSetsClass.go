package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type replicaSetListTypeAlias = appsv1.ReplicaSetList

//go:generate gotemplate "./templates/resource" "replicaSetsClass(\"ReplicaSets\", replicaSets, replicaSetListTypeAlias)"

func (c *replicaSetsClass) getList(ns string, listOptions metav1.ListOptions) (*appsv1.ReplicaSetList, error) {
	return c.rk.clientset.Apps().ReplicaSets(ns).List(listOptions)
}

func (c *replicaSetsClass) getItem(replicaSets replicaSetListTypeAlias, index int) (*replicaSetClassInstance, error) {
	newReplicaSetObj, err := c.rk.classes.ReplicaSet.New()
	if err != nil {
		return nil, err
	}
	replicaSet := replicaSets.Items[index]
	newReplicaSetObj.vars.replicaSet = replicaSetTypeAlias(replicaSet)
	return newReplicaSetObj, nil
}

//go:generate gotemplate "./templates/resource/list" "replicaSetsListModule(replicaSetsClass, \"ReplicaSets\", replicaSets, replicaSetListTypeAlias)"

func (c *replicaSetsClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *replicaSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
