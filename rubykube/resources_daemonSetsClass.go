package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type daemonSetListTypeAlias kext.DaemonSetList

//go:generate gotemplate "./templates/resource" "daemonSetsClass(\"DaemonSets\", daemonSets, daemonSetListTypeAlias)"

func (c *daemonSetsClass) getList(ns string, listOptions kapi.ListOptions) (*kext.DaemonSetList, error) {
	return c.rk.clientset.Extensions().DaemonSets(ns).List(listOptions)
}

func (c *daemonSetsClass) getItem(_ daemonSetListTypeAlias, _ int) (*podsClassInstance, error) {
	return nil, fmt.Errorf("Not implemented!")
}

//go:generate gotemplate "./templates/resource/list" "daemonSetsListModule(daemonSetsClass, \"DaemonSets\", daemonSets, daemonSetListTypeAlias)"

func (c *daemonSetsClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *daemonSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
