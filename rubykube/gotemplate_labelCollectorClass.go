package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type labelCollectorClass struct {
	class   *mruby.Class
	objects []labelCollectorClassInstance
	rk      *RubyKube
}

type labelCollectorClassInstance struct {
	self *mruby.MrbValue
	vars *labelCollectorClassInstanceVars
}

func newLabelCollectorClass(rk *RubyKube) *labelCollectorClass {
	c := &labelCollectorClass{objects: []labelCollectorClassInstance{}, rk: rk}
	c.class = defineLabelCollectorClass(rk, c)
	return c
}

func defineLabelCollectorClass(rk *RubyKube, c *labelCollectorClass) *mruby.Class {
	// common methods
	return rk.defineClass("LabelCollector", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *labelCollectorClass) New(args ...mruby.Value) (*labelCollectorClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newLabelCollectorClassInstanceVars(c, s, args...)
	o := labelCollectorClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelCollectorClass) LookupVars(this *mruby.MrbValue) (*labelCollectorClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "LabelCollector")
}
