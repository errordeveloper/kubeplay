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

type labelExpression struct {
	key      string
	operator string
	values   []string
}

type labelKeyClass struct {
	class   *mruby.Class
	objects []labelKeyInstance
	rk      *RubyKube
}

type labelKeyInstance struct {
	self *mruby.MrbValue
	vars *labelKeyInstanceVars
}

type matchFunc func(labelExpression)

type labelKeyInstanceVars struct {
	name    string
	onMatch matchFunc
}

type labelCollectorClass struct {
	class   *mruby.Class
	objects []labelCollectorInstance
	rk      *RubyKube
}

type labelCollectorInstance struct {
	self *mruby.MrbValue
	vars *labelCollectorInstanceVars
}

type labelCollectorInstanceVars struct {
	labels []labelExpression
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

func newLabelCollectorClass(rk *RubyKube) *labelCollectorClass {
	c := &labelCollectorClass{objects: []labelCollectorInstance{}, rk: rk}
	c.class = defineLabelCollectorClass(rk, c)
	return c
}

func defineLabelCollectorClass(rk *RubyKube, l *labelCollectorClass) *mruby.Class {
	return defineClass(rk, "LabelCollector", map[string]methodDefintion{
		"label": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := l.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				argv := []mruby.Value{}
				for _, arg := range m.GetArgs() {
					argv = append(argv, mruby.Value(arg))
				}

				newLabelKeyObj, err := rk.classes.LabelKey.New(argv...)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				// let it append to my lables
				newLabelKeyObj.vars.onMatch = func(e labelExpression) {
					vars.labels = append(vars.labels, e)
				}

				return newLabelKeyObj.self, nil
			},
			instanceMethod,
		},
	})
}

func (c *labelCollectorClass) New(args ...mruby.Value) (*labelCollectorInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	if args[0].MrbValue(c.rk.mrb).Type() != mruby.TypeProc {
		return nil, fmt.Errorf("Block must be given")
	}

	o := labelCollectorInstance{
		self: s,
		vars: &labelCollectorInstanceVars{
			[]labelExpression{},
		},
	}
	c.objects = append(c.objects, o)

	for _, v := range []string{
		"app",
		"name",
		"org",
		"owner",
		"project",
		"revision",
		"service",
		"team",
		"tier",
		"v",
		"version",
	} {
		// could do this, but it won't work cause we need to set onMatch somehow...
		// c.rk.mrb.LoadString(fmt.Sprintf("@%s = RubyKube::LabelKey.new(%s)", v, v))
		variableName, keyName := c.rk.mrb.StringValue("@"+v), c.rk.mrb.StringValue(v)

		l, err := c.rk.classes.LabelKey.New(keyName)
		if err != nil {
			return nil, err
		}

		l.vars.onMatch = func(e labelExpression) {
			o.vars.labels = append(o.vars.labels, e)
		}

		if _, err := s.Call("instance_variable_set", variableName, l.self); err != nil {
			return nil, err
		}
	}

	if _, err := s.CallBlock("instance_eval", args[0]); err != nil {
		return nil, err
	}

	return &o, nil
}

func (c *labelCollectorClass) LookupVars(this *mruby.MrbValue) (*labelCollectorInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func newLabelKeyClass(rk *RubyKube) *labelKeyClass {
	c := &labelKeyClass{objects: []labelKeyInstance{}, rk: rk}
	c.class = defineLabelKeyClass(rk, c)
	return c
}

func (l *labelKeyClass) makeMatchMethod(operator string) methodDefintion {
	return methodDefintion{
		mruby.ArgsReq(0) | mruby.ArgsOpt(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
			vars, err := l.LookupVars(self)
			if err != nil {
				return nil, createException(m, err.Error())
			}

			e := labelExpression{key: vars.name, operator: operator, values: []string{}}

			if err := l.appendSetExpression(&e.values, m.GetArgs()...); err != nil {
				return nil, createException(m, err.Error())
			}

			vars.onMatch(e)

			return nil, nil
		},
		instanceMethod,
	}
}

func defineLabelKeyClass(rk *RubyKube, l *labelKeyClass) *mruby.Class {
	return defineClass(rk, "LabelKey", map[string]methodDefintion{
		"=~":          l.makeMatchMethod("in"),
		"==":          l.makeMatchMethod("in"),
		"in":          l.makeMatchMethod("in"),
		"in?":         l.makeMatchMethod("in"),
		"is_in":       l.makeMatchMethod("in"),
		"is_in?":      l.makeMatchMethod("in"),
		"!~":          l.makeMatchMethod("notin"),
		"!=":          l.makeMatchMethod("notin"),
		"notin":       l.makeMatchMethod("notin"),
		"notin?":      l.makeMatchMethod("notin"),
		"not_in":      l.makeMatchMethod("notin"),
		"not_in?":     l.makeMatchMethod("notin"),
		"is_not_in":   l.makeMatchMethod("notin"),
		"is_not_in?":  l.makeMatchMethod("notin"),
		"any?":        l.makeMatchMethod(""),
		"is_set?":     l.makeMatchMethod(""),
		"defined?":    l.makeMatchMethod(""),
		"present?":    l.makeMatchMethod(""),
		"anything?":   l.makeMatchMethod(""),
		"is_present?": l.makeMatchMethod(""),
		"method_missing": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return nil, nil
			},
			instanceMethod},
	})
}

func (c *labelKeyClass) New(args ...mruby.Value) (*labelKeyInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("Exactly one argument must supplied")
	}

	o := labelKeyInstance{
		self: s,
		vars: &labelKeyInstanceVars{
			name: args[0].MrbValue(s.Mrb()).String(),
			onMatch: func(e labelExpression) {
				panic("Too early too call me!")
			},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *labelKeyClass) LookupVars(this *mruby.MrbValue) (*labelKeyInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (c *labelKeyClass) appendSetExpression(values *[]string, args ...*mruby.MrbValue) error {
	for _, m := range args {
		switch m.Type() {
		case mruby.TypeArray:
			if err := iterateArray(m, func(i int, v *mruby.MrbValue) error {
				if err := c.appendSetExpression(values, v); err != nil {
					return nil
				}
				return nil
			}); err != nil {
				return err
			}
		case mruby.TypeHash:
			return fmt.Errorf("a hash is an invalid type for label value set expression")
		default:
			s := m.String()
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("found an invalid string %q", s)
			}
			*values = append(*values, s)
		}
	}
	return nil
}
