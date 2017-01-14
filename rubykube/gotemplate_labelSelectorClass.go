package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type labelSelectorClass struct {
	class   *mruby.Class
	objects []labelSelectorClassInstance
	rk      *RubyKube
}

type labelSelectorClassInstance struct {
	self *mruby.MrbValue
	vars *labelSelectorClassInstanceVars
}

func newLabelSelectorClass(rk *RubyKube) *labelSelectorClass {
	c := &labelSelectorClass{objects: []labelSelectorClassInstance{}, rk: rk}
	c.class = defineLabelSelectorClass(rk, c)
	return c
}

func defineLabelSelectorClass(rk *RubyKube, c *labelSelectorClass) *mruby.Class {
	// common methods
	return rk.defineClass("LabelSelector", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *labelSelectorClass) New(args ...mruby.Value) (*labelSelectorClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newLabelSelectorClassInstanceVars(c, s, args...)
	o := labelSelectorClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelSelectorClass) LookupVars(this *mruby.MrbValue) (*labelSelectorClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "LabelSelector")
}
