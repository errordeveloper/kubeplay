package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

type fieldCollectorClassInstanceVars struct {
	fields []fieldExpression
	eval   func() error
}

func newFieldCollectorClassInstanceVars(c *fieldCollectorClass, s *mruby.MrbValue, args ...mruby.Value) (*fieldCollectorClassInstanceVars, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	if args[0].MrbValue(c.rk.mrb).Type() != mruby.TypeProc {
		return nil, fmt.Errorf("Block must be given")
	}

	o := &fieldCollectorClassInstanceVars{
		fields: []fieldExpression{},
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

		l.vars.onMatch = func(e labelExpression) { return }

		if _, err := s.Call("instance_variable_set", variableName, l.self); err != nil {
			return nil, err
		}
	}

	return o, nil
}

//go:generate gotemplate "./templates/basic" "fieldCollectorClass(\"FieldCollector\", newFieldCollectorClassInstanceVars, fieldCollectorClassInstanceVars)"

func (c *fieldCollectorClass) makeFieldMethod() methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := c.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			newFieldKeyObj, err := c.rk.classes.FieldKey.New(toValues(m.GetArgs())...)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			// let it append to my fields
			newFieldKeyObj.vars.onMatch = func(e fieldExpression) { vars.fields = append(vars.fields, e) }

			return newFieldKeyObj.self, nil
		},
		instanceMethod,
	}
}

func (c *fieldCollectorClass) makeLableMethod() methodDefintion {
	return methodDefintion{
		mruby.ArgsAny(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			newLabelKeyObj, err := c.rk.classes.LabelKey.New(toValues(m.GetArgs())...)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			newLabelKeyObj.vars.onMatch = func(_ labelExpression) { return }

			return newLabelKeyObj.self, nil
		},
		instanceMethod,
	}
}

func (c *fieldCollectorClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"field":          c.makeFieldMethod(),
		"label":          c.makeLableMethod(),
		"method_missing": c.makeFieldMethod(),
	})
}
