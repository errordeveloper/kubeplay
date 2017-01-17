package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type serviceListTypeAlias kapi.ServiceList

//go:generate gotemplate "./templates/resource" "servicesClass(\"Services\", services, serviceListTypeAlias)"

func (c *servicesClass) getList(ns string, listOptions kapi.ListOptions) (*kapi.ServiceList, error) {
	return c.rk.clientset.Core().Services(ns).List(listOptions)
}

func (c *servicesClass) getItem(_ serviceListTypeAlias, _ int) (*podsClassInstance, error) {
	return nil, fmt.Errorf("Not implemented!")
}

//go:generate gotemplate "./templates/resource/list" "servicesListModule(servicesClass, \"Services\", services, serviceListTypeAlias)"

func (c *servicesClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *servicesClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
