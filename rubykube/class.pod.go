package rubykube

import (
	"fmt"

	"github.com/errordeveloper/kubeplay/rubykube/converter"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podClass struct {
	class   *mruby.Class
	objects []podClassInstance
}

type podClassInstance struct {
	self *mruby.MrbValue
	vars *podClassInstanceVars
}

type podClassInstanceVars struct {
	pod *kapi.Pod
}

func newPodClass(rk *RubyKube) *podClass {
	c := &podClass{objects: []podClassInstance{}}
	c.class = definePodClass(rk, c)
	return c
}

func definePodClass(rk *RubyKube, p *podClass) *mruby.Class {
	return defineClass(rk, "Pod", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				vars.pod, err = rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).Get(vars.pod.ObjectMeta.Name)
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

				fmt.Printf("self: %s/%s\n", vars.pod.ObjectMeta.Namespace, vars.pod.ObjectMeta.Name)
				return self, nil
			},
			instanceMethod,
		},
		"delete!": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				if err = rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).Delete(vars.pod.ObjectMeta.Name, &kapi.DeleteOptions{}); err != nil {
					return nil, createException(m, err.Error())
				}

				return self, nil
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
				if err := rbconv.Convert(vars.pod); err != nil {
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

				return marshalToJSON(vars.pod, m)
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

func (c *podClass) New() (*podClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := podClassInstance{
		self: s,
		vars: &podClassInstanceVars{
			&kapi.Pod{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podClass) LookupVars(this *mruby.MrbValue) (*podClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *podClassInstance) Update() (mruby.Value, error) {
	v, err := o.self.Call("get!")
	if err != nil {
		return nil, err
	}
	return v, nil
}
