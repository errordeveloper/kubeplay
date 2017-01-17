package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(parentClass, classNameString, instanceVariableName, instanceVariableType)

type daemonSetsListModule struct{}

func (c *daemonSetsClass) defineListMethods() {
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

				daemonSets, err := c.getList(c.rk.GetNamespace(ns), *listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if nameRegexp != nil {
					for _, item := range daemonSets.Items {
						if nameRegexp.MatchString(item.ObjectMeta.Name) {
							vars.daemonSets.Items = append(vars.daemonSets.Items, item)
						}
					}
				} else {
					vars.daemonSets = daemonSetListTypeAlias(*daemonSets)
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

				for n, item := range vars.daemonSets.Items {
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

				return m.FixnumValue(len(vars.daemonSets.Items)), nil
			},
			instanceMethod,
		},
		/*
			"[]": {
				mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
					var pod kapi.Pod
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

					l := len(vars.pods.Items)

					if n.Fixnum() >= l {
						return nil, nil
					}

					if n.Fixnum() >= 0 {
						pod = vars.pods.Items[n.Fixnum()]
					} else if -(l-1) <= n.Fixnum() && n.Fixnum() < 0 {
						pod = vars.pods.Items[l+n.Fixnum()]
					} else {
						return nil, nil
					}
					//fmt.Printf("%d: %s/%s\n", n.Fixnum(), pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)

					newPodObj, err := c.rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = podTypeAlias(pod)
					return newPodObj.self, nil
				},
				instanceMethod,
			},
			"count": {
				mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
					vars, err := c.LookupVars(self)
					if err != nil {
						return nil, createException(m, err.Error())
					}

					return m.FixnumValue(len(vars.pods.Items)), nil
				},
				instanceMethod,
			},
			"first": {
				mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
					vars, err := c.LookupVars(self)
					if err != nil {
						return nil, createException(m, err.Error())
					}

					if len(vars.pods.Items) > 0 {
						newPodObj, err := c.rk.classes.Pod.New()
						if err != nil {
							return nil, createException(m, err.Error())
						}
						newPodObj.vars.pod = podTypeAlias(vars.pods.Items[0])
						return newPodObj.self, nil
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

					l := len(vars.pods.Items)
					if l > 0 {
						newPodObj, err := c.rk.classes.Pod.New()
						if err != nil {
							return nil, createException(m, err.Error())
						}
						newPodObj.vars.pod = podTypeAlias(vars.pods.Items[rand.Intn(l)])
						return newPodObj.self, nil
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

					l := len(vars.pods.Items)

					if l > 0 {
						newPodObj, err := c.rk.classes.Pod.New()
						if err != nil {
							return nil, createException(m, err.Error())
						}
						newPodObj.vars.pod = podTypeAlias(vars.pods.Items[l-1])
						return newPodObj.self, nil
					}
					return nil, nil
				},
				instanceMethod,
			},
		*/
	})
}
