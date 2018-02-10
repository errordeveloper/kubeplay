package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

type serviceListTypeAlias = corev1.ServiceList

//go:generate gotemplate "./templates/resource" "servicesClass(\"Services\", services, serviceListTypeAlias)"

func (c *servicesClass) getList(ns string, listOptions metav1.ListOptions) (*corev1.ServiceList, error) {
	return c.rk.clientset.Core().Services(ns).List(listOptions)
}

func (c *servicesClass) getItem(services serviceListTypeAlias, index int) (*serviceClassInstance, error) {
	newServiceObj, err := c.rk.classes.Service.New()
	if err != nil {
		return nil, err
	}
	service := services.Items[index]
	newServiceObj.vars.service = serviceTypeAlias(service)
	return newServiceObj, nil
}

//go:generate gotemplate "./templates/resource/list" "servicesListModule(servicesClass, \"Services\", services, serviceListTypeAlias)"

func (c *servicesClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *servicesClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
