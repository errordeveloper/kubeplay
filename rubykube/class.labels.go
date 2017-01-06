package rubykube

/*

import (
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

type labelName struct {
	class   *mruby.Class
	objects []labelNameInstance
}

type labelNameInstance struct {
	self *mruby.MrbValue
	vars *labelNameInstanceVars
}

type labelNameInstanceVars struct {
	key      string
	operator string
	values   []string
}

func newLabelNameClass(rk *RubyKube) *labelName {
	c := &labelName{objects: []labelNameInstance{}}
	c.class = defineLabelNameClass(rk, c)
	return c
}

func defineLabelNameClass(rk *RubyKube, p *labelName) *mruby.Class {
	return defineClass(rk, "LabelName", map[string]methodDefintion{})
}

func (c *labelName) New(args ...mruby.Value) (*labelNameInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := labelNameInstance{
		self: s,
		vars: &labelNameInstanceVars{
		//key: args[0].MrbValue(s.Mrb()).String(),
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelName) LookupVars(this *mruby.MrbValue) (*labelNameInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}
*/
