package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

type serviceTypeAlias = corev1.Service

//go:generate gotemplate "./templates/resource" "serviceClass(\"Service\", service, serviceTypeAlias)"

func (c *serviceClass) getSingleton(ns, name string) (*corev1.Service, error) {
	return c.rk.clientset.Core().Services(ns).Get(name, metav1.GetOptions{})
}

//go:generate gotemplate "./templates/resource/singleton" "serviceSingletonModule(serviceClass, \"Service\", service, serviceTypeAlias)"

//go:generate gotemplate "./templates/resource/podfinder" "servicePodFinderModule(serviceClass, \"Service\", service, serviceTypeAlias)"

func (c *serviceClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.definePodFinderMethods()
}

func (o *serviceClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
