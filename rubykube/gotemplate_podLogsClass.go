package rubykube

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

// template type RubyKubeClass(classNameString, newClassInstanceVars, classInstanceVarsType)

type podLogsClass struct {
	class   *mruby.Class
	objects []podLogsClassInstance
	rk      *RubyKube
}

type podLogsClassInstance struct {
	self *mruby.MrbValue
	vars *podLogsClassInstanceVars
}

func newPodLogsClass(rk *RubyKube) *podLogsClass {
	c := &podLogsClass{objects: []podLogsClassInstance{}, rk: rk}
	c.class = definePodLogsClass(rk, c)
	return c
}

func definePodLogsClass(rk *RubyKube, c *podLogsClass) *mruby.Class {
	// common methods
	return rk.defineClass("PodLogs", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(c.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *podLogsClass) New(args ...mruby.Value) (*podLogsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	v, err := newPodLogsClassInstanceVars(c, s, args...)
	if err != nil {
		return nil, err
	}

	o := podLogsClassInstance{
		self: s,
		vars: v,
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podLogsClass) LookupVars(this *mruby.MrbValue) (*podLogsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("%s: could not find class instance", "PodLogs")
}
