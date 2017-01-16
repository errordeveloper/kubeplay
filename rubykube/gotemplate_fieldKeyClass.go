package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type fieldKeyClass struct {
	class   *mruby.Class
	objects []fieldKeyClassInstance
	rk      *RubyKube
}

type fieldKeyClassInstance struct {
	self *mruby.MrbValue
	vars *fieldKeyClassInstanceVars
}

func newFieldKeyClass(rk *RubyKube) *fieldKeyClass {
	c := &fieldKeyClass{objects: []fieldKeyClassInstance{}, rk: rk}
	c.class = defineFieldKeyClass(rk, c)
	return c
}

func defineFieldKeyClass(rk *RubyKube, c *fieldKeyClass) *mruby.Class {
	// common methods
	return rk.defineClass("FieldKey", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *fieldKeyClass) New(args ...mruby.Value) (*fieldKeyClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newFieldKeyClassInstanceVars(c, s, args...)
	if err != nil {
		return nil, err
	}

	o := fieldKeyClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *fieldKeyClass) LookupVars(this *mruby.MrbValue) (*fieldKeyClassInstanceVars, error) {
	fmt.Printf("c.objects=%v\n", c.objects)
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "FieldKey")
}
