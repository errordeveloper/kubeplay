package resourcesingleton

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type RubyKubeClass struct{}

type parentClass int
type classNameString string
type instanceVariableName int
type instanceVariableType int

func (c *parentClass) defineSingletonMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				meta := vars.instanceVariableName.ObjectMeta
				instanceVariableName, err := c.getSignleton(meta.Namespace, meta.Name)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				vars.instanceVariableName = instanceVariableType(*instanceVariableName)
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

				fmt.Printf("self: %s/%s\n", vars.instanceVariableName.ObjectMeta.Namespace, vars.instanceVariableName.ObjectMeta.Name)
				return self, nil
			},
			instanceMethod,
		},
	})
}
