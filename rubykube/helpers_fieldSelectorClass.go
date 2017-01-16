package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
)

type fieldSelectorClassInstanceVars struct {
	collector *fieldCollectorClassInstance
}

func newFieldSelectorClassInstanceVars(c *fieldSelectorClass, s *mruby.MrbValue, args ...mruby.Value) (*fieldSelectorClassInstanceVars, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	newFieldCollectorObj, err := c.rk.classes.FieldCollector.New(args...)
	if err != nil {
		return nil, err
	}

	if err := newFieldCollectorObj.vars.eval(); err != nil {
		return nil, err
	}

	return &fieldSelectorClassInstanceVars{newFieldCollectorObj}, nil
}

//go:generate gotemplate "./templates/basic" "fieldSelectorClass(\"FieldSelector\", newFieldSelectorClassInstanceVars, fieldSelectorClassInstanceVars)"

func (c *fieldSelectorClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"to_s": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				fields := []string{}
				for _, e := range vars.collector.vars.fields {
					for _, v := range e.values {
						fields = append(fields, fmt.Sprintf("%s%s%s", e.key, e.operator, v))
					}
				}

				return m.StringValue(strings.Join(fields, ",")), nil
			},
			instanceMethod,
		},
	})
}
