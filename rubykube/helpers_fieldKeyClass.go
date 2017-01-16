package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
)

type fieldExpression struct {
	key      string
	operator string
	values   []string
}

type fieldMatchFunc func(fieldExpression)

type fieldKeyClassInstanceVars struct {
	name    []string
	onMatch fieldMatchFunc
}

func newFieldKeyClassInstanceVars(c *fieldKeyClass, s *mruby.MrbValue, args ...mruby.Value) (*fieldKeyClassInstanceVars, error) {
	o := &fieldKeyClassInstanceVars{
		name: []string{},
		onMatch: func(e fieldExpression) {
			panic("Too early too call me!")
		},
	}

	if len(args) > 0 {
		for _, v := range args {
			o.name = append(o.name, v.MrbValue(s.Mrb()).String())
		}
	}

	return o, nil
}

//go:generate gotemplate "./templates/basic" "fieldKeyClass(\"FieldKey\", newFieldKeyClassInstanceVars, fieldKeyClassInstanceVars)"

func (c *fieldKeyClass) makeMatchMethod(operator string) methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(0) | mruby.ArgsOpt(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := c.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			e := fieldExpression{key: strings.Join(vars.name, "."), operator: operator, values: []string{}}

			if err := c.appendSetExpression(&e.values, m.GetArgs()...); err != nil {
				return nil, createException(m, err.Error())
			}

			vars.onMatch(e)

			return nil, nil
		},
		instanceMethod,
	}
}

func (c *fieldKeyClass) appendSetExpression(values *[]string, args ...*mruby.MrbValue) error {
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
			return fmt.Errorf("a hash is an invalid type for field value expression")
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

func (c *fieldKeyClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"=~": c.makeMatchMethod("=="),
		"==": c.makeMatchMethod("=="),
		"!~": c.makeMatchMethod("!="),
		"!=": c.makeMatchMethod("!="),
		"to_s": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.StringValue(strings.Join(vars.name, ".")), nil
			},
			instanceMethod,
		},
		"method_missing": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				value, err := c.class.New(toValues(m.GetArgs())...)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return value, nil
			},
			instanceMethod,
		},
	})
}
