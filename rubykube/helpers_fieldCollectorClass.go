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

	return o, nil
}

//go:generate gotemplate "./templates/basic" "fieldCollectorClass(\"FieldCollector\", newFieldCollectorClassInstanceVars, fieldCollectorClassInstanceVars)"

func (c *fieldCollectorClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"field": {
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
				newFieldKeyObj.vars.onMatch = func(e fieldExpression) {
					vars.fields = append(vars.fields, e)
				}

				return newFieldKeyObj.self, nil
			},
			instanceMethod,
		},
	})
}
