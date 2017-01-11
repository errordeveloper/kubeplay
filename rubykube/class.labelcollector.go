package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

type labelCollectorClass struct {
	class   *mruby.Class
	objects []labelCollectorInstance
	rk      *RubyKube
}

type labelCollectorInstance struct {
	self *mruby.MrbValue
	vars *labelCollectorInstanceVars
}

type labelCollectorInstanceVars struct {
	labels []labelExpression
}

func newLabelCollectorClass(rk *RubyKube) *labelCollectorClass {
	c := &labelCollectorClass{objects: []labelCollectorInstance{}, rk: rk}
	c.class = defineLabelCollectorClass(rk, c)
	return c
}

func defineLabelCollectorClass(rk *RubyKube, l *labelCollectorClass) *mruby.Class {
	return defineClass(rk, "LabelCollector", map[string]methodDefintion{
		"label": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := l.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				newLabelKeyObj, err := rk.classes.LabelKey.New(toValues(m.GetArgs())...)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				// let it append to my lables
				newLabelKeyObj.vars.onMatch = func(e labelExpression) {
					vars.labels = append(vars.labels, e)
				}

				return newLabelKeyObj.self, nil
			},
			instanceMethod,
		},
	})
}

func (c *labelCollectorClass) New(args ...mruby.Value) (*labelCollectorInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	if args[0].MrbValue(c.rk.mrb).Type() != mruby.TypeProc {
		return nil, fmt.Errorf("Block must be given")
	}

	o := labelCollectorInstance{
		self: s,
		vars: &labelCollectorInstanceVars{
			[]labelExpression{},
		},
	}
	c.objects = append(c.objects, o)

	for _, v := range []string{
		"app",
		"name",
		"org",
		"owner",
		"project",
		"revision",
		"service",
		"team",
		"tier",
		"v",
		"version",
	} {
		// could do this, but it won't work cause we need to set onMatch somehow...
		// c.rk.mrb.LoadString(fmt.Sprintf("@%s = RubyKube::LabelKey.new(%s)", v, v))
		variableName, keyName := c.rk.mrb.StringValue("@"+v), c.rk.mrb.StringValue(v)

		l, err := c.rk.classes.LabelKey.New(keyName)
		if err != nil {
			return nil, err
		}

		l.vars.onMatch = func(e labelExpression) {
			o.vars.labels = append(o.vars.labels, e)
		}

		if _, err := s.Call("instance_variable_set", variableName, l.self); err != nil {
			return nil, err
		}
	}

	if _, err := s.CallBlock("instance_eval", args[0]); err != nil {
		return nil, err
	}

	return &o, nil
}

func (c *labelCollectorClass) LookupVars(this *mruby.MrbValue) (*labelCollectorInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}
