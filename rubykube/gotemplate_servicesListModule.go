package rubykube

import (
	"fmt"
	"math/rand"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type servicesListModule struct{}

func (c *servicesClass) defineListMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0) | mruby.ArgsOpt(2), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns, nameRegexp, listOptions, err := c.rk.resourceArgs(m.GetArgs())
				if err != nil {
					return nil, createException(m, err.Error())
				}

				services, err := c.getList(c.rk.GetNamespace(ns), *listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if nameRegexp != nil {
					for _, item := range services.Items {
						if nameRegexp.MatchString(item.ObjectMeta.Name) {
							vars.services.Items = append(vars.services.Items, item)
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

				for n, item := range vars.services.Items {
					fmt.Printf("%d: %s/%s\n", n, item.ObjectMeta.Namespace, item.ObjectMeta.Name)
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
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()
				err = standardCheck(c.rk, args, 1)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				n := args[0]
				if n.Type() != mruby.TypeFixnum {
					return nil, createException(m, "Argument must be an integer")
				}

				l := len(vars.services.Items)
				i := n.Fixnum()

				if i >= l {
					return nil, nil
				}

				if i < 0 {
					// handle negative index in the way Ruby does it, i.e. no infinit wrapping
					if -i <= l {
						i %= l
						i *= -1 // in Go, unlike Ruby this needs to be converted to positive value
					} else {
						return nil, nil
					}
				}

				obj, err := c.getItem(vars.services, i)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return obj.self, nil
			},
			instanceMethod,
		},
		"first": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if len(vars.services.Items) > 0 {
					obj, err := c.getItem(vars.services, 0)
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"any": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.services.Items)
				if l > 0 {
					obj, err := c.getItem(vars.services, rand.Intn(l))
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"last": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.services.Items)

				if l > 0 {
					obj, err := c.getItem(vars.services, l-1)
					if err != nil {
						return nil, createException(m, err.Error())
					}
					return obj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
	})
}
