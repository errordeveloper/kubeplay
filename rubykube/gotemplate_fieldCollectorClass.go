package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type fieldCollectorClass struct {
	class   *mruby.Class
	objects []fieldCollectorClassInstance
	rk      *RubyKube
}

type fieldCollectorClassInstance struct {
	self *mruby.MrbValue
	vars *fieldCollectorClassInstanceVars
}

func newFieldCollectorClass(rk *RubyKube) *fieldCollectorClass {
	c := &fieldCollectorClass{objects: []fieldCollectorClassInstance{}, rk: rk}
	c.class = defineFieldCollectorClass(rk, c)
	return c
}

func defineFieldCollectorClass(rk *RubyKube, c *fieldCollectorClass) *mruby.Class {
	// common methods
	return rk.defineClass("FieldCollector", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *fieldCollectorClass) New(args ...mruby.Value) (*fieldCollectorClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newFieldCollectorClassInstanceVars(c, s, args...)
	if err != nil {
		return nil, err
	}

	o := fieldCollectorClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *fieldCollectorClass) LookupVars(this *mruby.MrbValue) (*fieldCollectorClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "FieldCollector")
}
