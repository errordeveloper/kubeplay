package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
)

type labelSelectorClassInstanceVars struct {
	collector *labelCollectorClassInstance
}

func newLabelSelectorClassInstanceVars(c *labelSelectorClass, s *mruby.MrbValue, args ...mruby.Value) (*labelSelectorClassInstanceVars, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	newLabelCollectorObj, err := c.rk.classes.LabelCollector.New(args...)
	if err != nil {
		return nil, err
	}

	if err := newLabelCollectorObj.vars.eval(); err != nil {
		return nil, err
	}

	return &labelSelectorClassInstanceVars{newLabelCollectorObj}, nil
}

//go:generate gotemplate "./templates/basic" "labelSelectorClass(\"LabelSelector\", newLabelSelectorClassInstanceVars, labelSelectorClassInstanceVars)"

func (c *labelSelectorClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"to_s": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				labels := []string{}
				for _, e := range vars.collector.vars.labels {
					if e.operator != "" {
						labels = append(labels, fmt.Sprintf("%s %s (%s)", e.key, e.operator, strings.Join(e.values, ", ")))
					} else {
						labels = append(labels, e.key)
					}
				}

				return m.StringValue(strings.Join(labels, ",")), nil
			},
			instanceMethod,
		},
	})
}
