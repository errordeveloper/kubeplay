package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type labelKeyClass struct {
	class   *mruby.Class
	objects []labelKeyClassInstance
	rk      *RubyKube
}

type labelKeyClassInstance struct {
	self *mruby.MrbValue
	vars *labelKeyClassInstanceVars
}

func newLabelKeyClass(rk *RubyKube) *labelKeyClass {
	c := &labelKeyClass{objects: []labelKeyClassInstance{}, rk: rk}
	c.class = defineLabelKeyClass(rk, c)
	return c
}

func defineLabelKeyClass(rk *RubyKube, c *labelKeyClass) *mruby.Class {
	// common methods
	return rk.defineClass("LabelKey", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *labelKeyClass) New(args ...mruby.Value) (*labelKeyClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newLabelKeyClassInstanceVars(c, s, args...)
	o := labelKeyClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelKeyClass) LookupVars(this *mruby.MrbValue) (*labelKeyClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "LabelKey")
}
