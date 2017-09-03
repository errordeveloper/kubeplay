package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type serviceSingletonModule struct{}

func (c *serviceClass) defineSingletonMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				meta := vars.service.ObjectMeta
				service, err := c.getSingleton(meta.Namespace, meta.Name)
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
