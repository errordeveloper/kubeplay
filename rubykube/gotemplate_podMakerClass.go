package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type podMakerClass struct {
	class   *mruby.Class
	objects []podMakerClassInstance
	rk      *RubyKube
}

type podMakerClassInstance struct {
	self *mruby.MrbValue
	vars *podMakerClassInstanceVars
}

func newPodMakerClass(rk *RubyKube) *podMakerClass {
	c := &podMakerClass{objects: []podMakerClassInstance{}, rk: rk}
	c.class = definePodMakerClass(rk, c)
	return c
}

func definePodMakerClass(rk *RubyKube, c *podMakerClass) *mruby.Class {
	// common methods
	return rk.defineClass("PodMaker", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *podMakerClass) New(args ...mruby.Value) (*podMakerClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newPodMakerClassInstanceVars(c, s, args...)
	o := podMakerClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podMakerClass) LookupVars(this *mruby.MrbValue) (*podMakerClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "PodMaker")
}
