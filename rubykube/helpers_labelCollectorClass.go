package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

type labelCollectorClassInstanceVars struct {
	labels []labelExpression
	eval   func() error
}

var wellKnownLabels = []string{
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
}

func newLabelCollectorClassInstanceVars(c *labelCollectorClass, s *mruby.MrbValue, args ...mruby.Value) (*labelCollectorClassInstanceVars, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	if args[0].MrbValue(c.rk.mrb).Type() != mruby.TypeProc {
		return nil, fmt.Errorf("Block must be given")
	}

	o := &labelCollectorClassInstanceVars{
		labels: []labelExpression{},
		eval: func() error {
			if _, err := s.CallBlock("instance_eval", args[0]); err != nil {
				return err
			}
			return nil
		},
	}

	for _, v := range wellKnownLabels {
		// could do this, but it won't work cause we need to set onMatch somehow...
		// c.rk.mrb.LoadString(fmt.Sprintf("@%s = RubyKube::LabelKey.new(%s)", v, v))
		variableName, keyName := c.rk.mrb.StringValue("@"+v), c.rk.mrb.StringValue(v)

		l, err := c.rk.classes.LabelKey.New(keyName)
		if err != nil {
			return nil, err
		}

		l.vars.onMatch = func(e labelExpression) {
			o.labels = append(o.labels, e)
		}

		if _, err := s.Call("instance_variable_set", variableName, l.self); err != nil {
			return nil, err
		}
	}

	return o, nil
}

//go:generate gotemplate "./templates/basic" "labelCollectorClass(\"LabelCollector\", newLabelCollectorClassInstanceVars, labelCollectorClassInstanceVars)"

func (c *labelCollectorClass) makeLableMethod() methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := c.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			newLabelKeyObj, err := c.rk.classes.LabelKey.New(toValues(m.GetArgs())...)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			// let it append to my labels
			newLabelKeyObj.vars.onMatch = func(e labelExpression) {
				vars.labels = append(vars.labels, e)
			}

			return newLabelKeyObj.self, nil
		},
		instanceMethod,
	}
}

func (c *labelCollectorClass) makeFieldMethod() methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			newFieldKeyObj, err := c.rk.classes.FieldKey.New(toValues(m.GetArgs())...)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			newFieldKeyObj.vars.onMatch = func(_ fieldExpression) { return }

			return newFieldKeyObj.self, nil
		},
		instanceMethod,
	}
}

func (c *labelCollectorClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"field":          c.makeFieldMethod(),
		"label":          c.makeLableMethod(),
		"method_missing": c.makeLableMethod(),
	})
}
