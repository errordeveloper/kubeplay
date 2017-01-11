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

type labelKeyClass struct {
	class   *mruby.Class
	objects []labelKeyInstance
	rk      *RubyKube
}

type labelKeyInstance struct {
	self *mruby.MrbValue
	vars *labelKeyInstanceVars
}

type matchFunc func(labelExpression)

type labelKeyInstanceVars struct {
	name    string
	onMatch matchFunc
}

func (l *labelKeyClass) makeMatchMethod(operator string) methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(0) | mruby.ArgsOpt(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := l.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			e := labelExpression{key: vars.name, operator: operator, values: []string{}}

			if err := l.appendSetExpression(&e.values, m.GetArgs()...); err != nil {
				return nil, createException(m, err.Error())
			}

			vars.onMatch(e)

			return nil, nil
		},
		instanceMethod,
	}
}

func newLabelKeyClass(rk *RubyKube) *labelKeyClass {
	c := &labelKeyClass{objects: []labelKeyInstance{}, rk: rk}
	c.class = defineLabelKeyClass(rk, c)
	return c
}

func defineLabelKeyClass(rk *RubyKube, l *labelKeyClass) *mruby.Class {
	return defineClass(rk, "LabelKey", map[string]methodDefintion{
		"=~":          l.makeMatchMethod("in"),
		"==":          l.makeMatchMethod("in"),
		"in":          l.makeMatchMethod("in"),
		"in?":         l.makeMatchMethod("in"),
		"is_in":       l.makeMatchMethod("in"),
		"is_in?":      l.makeMatchMethod("in"),
		"!~":          l.makeMatchMethod("notin"),
		"!=":          l.makeMatchMethod("notin"),
		"notin":       l.makeMatchMethod("notin"),
		"notin?":      l.makeMatchMethod("notin"),
		"not_in":      l.makeMatchMethod("notin"),
		"not_in?":     l.makeMatchMethod("notin"),
		"is_not_in":   l.makeMatchMethod("notin"),
		"is_not_in?":  l.makeMatchMethod("notin"),
		"any?":        l.makeMatchMethod(""),
		"is_set?":     l.makeMatchMethod(""),
		"defined?":    l.makeMatchMethod(""),
		"present?":    l.makeMatchMethod(""),
		"anything?":   l.makeMatchMethod(""),
		"is_present?": l.makeMatchMethod(""),
		"method_missing": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return nil, nil
			},
			instanceMethod},
	})
}

func (c *labelKeyClass) New(args ...mruby.Value) (*labelKeyInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	o := labelKeyInstance{
		self: s,
		vars: &labelKeyInstanceVars{
			name: args[0].MrbValue(s.Mrb()).String(),
			onMatch: func(e labelExpression) {
				panic("Too early too call me!")
			},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelKeyClass) LookupVars(this *mruby.MrbValue) (*labelKeyInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
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
