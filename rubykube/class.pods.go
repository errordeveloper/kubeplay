package rubykube

import (
	"fmt"
	"math/rand"
	"regexp"

	"github.com/errordeveloper/kubeplay/rubykube/converter"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podsClass struct {
	class   *mruby.Class
	objects []podsClassInstance
}

type podsClassInstance struct {
	self *mruby.MrbValue
	vars *podsClassInstanceVars
}

type podsClassInstanceVars struct {
	pods *kapi.PodList
}

func newPodsClass(rk *RubyKube) *podsClass {
	c := &podsClass{objects: []podsClassInstance{}}
	c.class = definePodsClass(rk, c)
	return c
}

func definePodsClass(rk *RubyKube, p *podsClass) *mruby.Class {
	return defineClass(rk, "Pods", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				var (
					ns          string
					listOptions kapi.ListOptions
				)

				args := m.GetArgs()
				fmt.Printf("args=%+v\n", args)

				validName := "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

				// pods "foo/"
				// pods "*/"
				podsWithinNamespace := regexp.MustCompile(fmt.Sprintf(`^(?P<namespace>%s|\*)\/$`, validName))

				// pods "foo/bar-*"
				// pods "bar-*"
				// pods "*/bar-*"

				hasNameGlob := false
				hasSelectors := false

				parseNameGlob := func(arg *mruby.MrbValue) error {
					s := arg.String()
					switch {
					case podsWithinNamespace.MatchString(s):
						getNamedMatch(podsWithinNamespace, s, "namespace", &ns)
					default:
						fmt.Printf("Invalid glob expression - try `pods \"<namespace>/\"`, `pods \"*/\"` or `pods \"*/foo-*\"`")
					}

					hasNameGlob = true
					return nil
				}

				parseSelectors := func(arg *mruby.MrbValue) error {
					pc, err := NewParamsCollection(arg,
						params{
							allowed:   []string{"labels", "fields"},
							required:  []string{},
							valueType: mruby.TypeString,
						},
					)

					if err != nil {
						return err
					}

					p := pc.ToMapOfStrings()

					if v, ok := p["labels"]; ok {
						listOptions.LabelSelector = v
					}
					if v, ok := p["fields"]; ok {
						listOptions.FieldSelector = v
					}

					hasSelectors = true
					return nil
				}

				if len(args) >= 1 {
					switch args[0].Type() {
					case mruby.TypeString:
						if err := parseNameGlob(args[0]); err != nil {
							return nil, createException(m, err.Error())
						}
					case mruby.TypeHash:
						if err := parseSelectors(args[0]); err != nil {
							return nil, createException(m, err.Error())
						}
					case mruby.TypeArray:
						// TODO: we could allow users to collect object matching multiple globs
						return nil, createException(m, "Not yet implemented!")
					}
				}

				if len(args) >= 2 {
					secondArgError := func(kind string) (mruby.Value, mruby.Value) {
						return nil, createException(m, "Found second "+kind+" argument, only single one is allowed - use array notation for mulptiple queries")
					}

					switch args[1].Type() {
					case mruby.TypeString:
						if hasNameGlob {
							return secondArgError("glob")
						}
						if err := parseNameGlob(args[1]); err != nil {
							return nil, createException(m, err.Error())
						}
					case mruby.TypeHash:
						if hasSelectors {
							return secondArgError("selector")
						}
						if err := parseSelectors(args[1]); err != nil {
							return nil, createException(m, err.Error())
						}
					case mruby.TypeArray:
						return nil, createException(m, "Only single array argument is allowed")
					}
				}

				if len(args) >= 3 {
					return nil, createException(m, "Maximum 2 arguments allowed")
				}

				vars.pods, err = rk.clientset.Core().Pods(rk.GetNamespace(ns)).List(listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return self, nil
			},
			instanceMethod,
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				for n, pod := range vars.pods.Items {
					fmt.Printf("%d: %s/%s\n", n, pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
				}
				return self, nil
			},
			instanceMethod,
		},
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				var pod kapi.Pod
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()
				err = standardCheck(rk, args, 1)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				n := args[0]
				if n.Type() != mruby.TypeFixnum {
					return nil, createException(m, "Argument must be an integer")
				}

				l := len(vars.pods.Items)

				if n.Fixnum() >= l {
					return nil, nil
				}

				if n.Fixnum() >= 0 {
					pod = vars.pods.Items[n.Fixnum()]
				} else if -(l-1) <= n.Fixnum() && n.Fixnum() < 0 {
					pod = vars.pods.Items[l+n.Fixnum()]
				} else {
					return nil, nil
				}
				//fmt.Printf("%d: %s/%s\n", n.Fixnum(), pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)

				newPodObj, err := rk.classes.Pod.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				newPodObj.vars.pod = &pod
				return newPodObj.self, nil
			},
			instanceMethod,
		},
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				rbconv := converter.New(m)
				if err := rbconv.Convert(vars.pods); err != nil {
					return nil, createException(m, err.Error())
				}

				return rbconv.Value(), nil
			},
			instanceMethod,
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return marshalToJSON(vars.pods, m)
			},
			instanceMethod,
		},
		"count": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return m.FixnumValue(len(vars.pods.Items)), nil
			},
			instanceMethod,
		},
		"first": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if len(vars.pods.Items) > 0 {
					newPodObj, err := rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = &vars.pods.Items[0]
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"any": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.pods.Items)
				if l > 0 {
					newPodObj, err := rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = &vars.pods.Items[rand.Intn(l)]
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"last": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				l := len(vars.pods.Items)

				if l > 0 {
					newPodObj, err := rk.classes.Pod.New()
					if err != nil {
						return nil, createException(m, err.Error())
					}
					newPodObj.vars.pod = &vars.pods.Items[l-1]
					return newPodObj.self, nil
				}
				return nil, nil
			},
			instanceMethod,
		},
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(p.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *podsClass) New() (*podsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := podsClassInstance{
		self: s,
		vars: &podsClassInstanceVars{
			&kapi.PodList{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podsClass) LookupVars(this *mruby.MrbValue) (*podsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *podsClassInstance) Update(args ...mruby.Value) (mruby.Value, error) {
	v, err := o.self.Call("get!", args...)
	if err != nil {
		return nil, err
	}
	return v, nil
}
