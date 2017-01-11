package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
)

type labelSelectorClass struct {
	class   *mruby.Class
	objects []labelSelectorInstance
	rk      *RubyKube // TODO copy to other classes
}

type labelSelectorInstance struct {
	self *mruby.MrbValue
	vars *labelSelectorInstanceVars
}

type labelSelectorInstanceVars struct {
	collector *labelCollectorInstance
}

func newLabelSelectorClass(rk *RubyKube) *labelSelectorClass {
	c := &labelSelectorClass{objects: []labelSelectorInstance{}, rk: rk}
	c.class = defineLabelSelectorClass(rk, c)
	return c
}

func defineLabelSelectorClass(rk *RubyKube, l *labelSelectorClass) *mruby.Class {
	return defineClass(rk, "LabelSelector", map[string]methodDefintion{
		"to_s": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := l.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				labels := []string{}
				for _, e := range vars.collector.vars.labels {
					if e.operator != "" {
						labels = append(labels, fmt.Sprintf("%s %s (%s)", e.key, e.operator, strings.Join(e.values, ", ")))
					} else {
						labels = append(labels, e.key)
					}
				}

				return m.StringValue(strings.Join(labels, ",")), nil
			},
			instanceMethod,
		},
	})
}

func (c *labelSelectorClass) New(args ...mruby.Value) (*labelSelectorInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	//if args[0].MrbValue(m).Type() != mruby.TypeProc {
	//	return nil, fmt.Errorf("Block must be given")
	//}

	newLabelCollectorObj, err := c.rk.classes.LabelCollector.New(args...)
	if err != nil {
		return nil, err
	}

	o := labelSelectorInstance{
		self: s,
		vars: &labelSelectorInstanceVars{
			newLabelCollectorObj,
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil

}
func (c *labelSelectorClass) LookupVars(this *mruby.MrbValue) (*labelSelectorInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}
