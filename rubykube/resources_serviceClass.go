package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type serviceTypeAlias kapi.Service

//go:generate gotemplate "./templates/resource" "serviceClass(\"Service\", service, serviceTypeAlias)"

func (c *serviceClass) getSignleton(ns, name string) (*kapi.Service, error) {
	return c.rk.clientset.Core().Services(ns).Get(name)
}

//go:generate gotemplate "./templates/resource/singleton" "serviceSignletonModule(serviceClass, \"Service\", service, serviceTypeAlias)"

func (c *serviceClass) defineOwnMethods() {
	c.defineSingletonMethods()
}

func (o *serviceClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
