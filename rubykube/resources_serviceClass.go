package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type serviceTypeAlias kapi.Service

//go:generate gotemplate "./templates/resource" "serviceClass(\"Service\", service, serviceTypeAlias)"

func (c *serviceClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				service, err := c.rk.clientset.Core().Services(vars.service.ObjectMeta.Namespace).Get(vars.service.ObjectMeta.Name)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				vars.service = serviceTypeAlias(*service)
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

				fmt.Printf("self: %s/%s\n", vars.service.ObjectMeta.Namespace, vars.service.ObjectMeta.Name)
				return self, nil
			},
			instanceMethod,
		},
	})
}

func (o *serviceClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
