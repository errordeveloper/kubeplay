package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type fieldSelectorClass struct {
	class   *mruby.Class
	objects []fieldSelectorClassInstance
	rk      *RubyKube
}

type fieldSelectorClassInstance struct {
	self *mruby.MrbValue
	vars *fieldSelectorClassInstanceVars
}

func newFieldSelectorClass(rk *RubyKube) *fieldSelectorClass {
	c := &fieldSelectorClass{objects: []fieldSelectorClassInstance{}, rk: rk}
	c.class = defineFieldSelectorClass(rk, c)
	return c
}

func defineFieldSelectorClass(rk *RubyKube, c *fieldSelectorClass) *mruby.Class {
	// common methods
	return rk.defineClass("FieldSelector", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *fieldSelectorClass) New(args ...mruby.Value) (*fieldSelectorClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newFieldSelectorClassInstanceVars(c, s, args...)
	if err != nil {
		return nil, err
	}

	o := fieldSelectorClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *fieldSelectorClass) LookupVars(this *mruby.MrbValue) (*fieldSelectorClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "FieldSelector")
}
