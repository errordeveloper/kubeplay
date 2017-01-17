package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type daemonSetSingletonModule struct{}

func (c *daemonSetClass) defineSingletonMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				meta := vars.daemonSet.ObjectMeta
				daemonSet, err := c.getSignleton(meta.Namespace, meta.Name)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				vars.daemonSet = daemonSetTypeAlias(*daemonSet)
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

				fmt.Printf("self: %s/%s\n", vars.daemonSet.ObjectMeta.Namespace, vars.daemonSet.ObjectMeta.Name)
				return self, nil
			},
			instanceMethod,
		},
	})
}
