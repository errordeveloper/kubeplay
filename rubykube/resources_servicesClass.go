package rubykube

import (
	"fmt"
	_ "math/rand"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type serviceListTypeAlias kapi.ServiceList

//go:generate gotemplate "./templates/resource" "servicesClass(\"Services\", services, serviceListTypeAlias)"

func (c *servicesClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, serviceNameRegexp, listOptions, err := c.rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				services, err := c.rk.clientset.Core().Services(c.rk.GetNamespace(ns)).List(*listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if serviceNameRegexp != nil {
					for _, service := range services.Items {
						if serviceNameRegexp.MatchString(service.ObjectMeta.Name) {
							vars.services.Items = append(vars.services.Items, service)
						}
					}
				} else {
					vars.services = serviceListTypeAlias(*services)
				}
				return self, nil
			},
			instanceMethod,
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				for n, service := range vars.services.Items {
					fmt.Printf("%d: %s/%s\n", n, service.ObjectMeta.Namespace, service.ObjectMeta.Name)
				}
				return self, nil
			},
			instanceMethod,
		},
		"count": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.FixnumValue(len(vars.services.Items)), nil
			},
			instanceMethod,
		},
	})
}

func (o *servicesClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
