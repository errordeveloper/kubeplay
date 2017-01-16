package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
)

type labelExpression struct {
	key      string
	operator string
	values   []string
}

type labelMatchFunc func(labelExpression)

type labelKeyClassInstanceVars struct {
	name    string
	onMatch labelMatchFunc
}

func newLabelKeyClassInstanceVars(c *labelKeyClass, s *mruby.MrbValue, args ...mruby.Value) (*labelKeyClassInstanceVars, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	o := &labelKeyClassInstanceVars{
		name: args[0].MrbValue(s.Mrb()).String(),
		onMatch: func(e labelExpression) {
			panic("Too early too call me!")
		},
	}
	return o, nil
}

//go:generate gotemplate "./templates/basic" "labelKeyClass(\"LabelKey\", newLabelKeyClassInstanceVars, labelKeyClassInstanceVars)"

func (c *labelKeyClass) makeMatchMethod(operator string) methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(0) | mruby.ArgsOpt(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := c.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			e := labelExpression{key: vars.name, operator: operator, values: []string{}}

			if err := c.appendSetExpression(&e.values, m.GetArgs()...); err != nil {
				return nil, createException(m, err.Error())
			}

			vars.onMatch(e)

			return nil, nil
		},
		instanceMethod,
	}
}

func (c *labelKeyClass) appendSetExpression(values *[]string, args ...*mruby.MrbValue) error {
	for _, m := range args {
		switch m.Type() {
		case mruby.TypeArray:
			if err := iterateArray(m, func(i int, v *mruby.MrbValue) error {
				if err := c.appendSetExpression(values, v); err != nil {
					return nil
				}
				return nil
			}); err != nil {
				return err
			}
		case mruby.TypeHash:
			return fmt.Errorf("a hash is an invalid type for label value set expression")
		default:
			s := m.String()
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("found an invalid string %q", s)
			}
			*values = append(*values, s)
		}
	}
	return nil
}

func (c *labelKeyClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"=~":          c.makeMatchMethod("in"),
		"==":          c.makeMatchMethod("in"),
		"in":          c.makeMatchMethod("in"),
		"in?":         c.makeMatchMethod("in"),
		"is_in":       c.makeMatchMethod("in"),
		"is_in?":      c.makeMatchMethod("in"),
		"!~":          c.makeMatchMethod("notin"),
		"!=":          c.makeMatchMethod("notin"),
		"notin":       c.makeMatchMethod("notin"),
		"notin?":      c.makeMatchMethod("notin"),
		"not_in":      c.makeMatchMethod("notin"),
		"not_in?":     c.makeMatchMethod("notin"),
		"is_not_in":   c.makeMatchMethod("notin"),
		"is_not_in?":  c.makeMatchMethod("notin"),
		"any?":        c.makeMatchMethod(""),
		"is_set?":     c.makeMatchMethod(""),
		"defined?":    c.makeMatchMethod(""),
		"present?":    c.makeMatchMethod(""),
		"anything?":   c.makeMatchMethod(""),
		"is_present?": c.makeMatchMethod(""),
		"method_missing": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return nil, nil
			},
			instanceMethod,
		},
	})
}
