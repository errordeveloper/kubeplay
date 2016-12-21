package rubykube

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

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

				vars.pods, err = rk.clientset.Core().Pods("").List(kapi.ListOptions{})
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
					return nil, createException(m, "Argument must be a integer")
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

func (o *podsClassInstance) Update() (mruby.Value, error) {
	v, err := o.self.Call("get!")
	if err != nil {
		return nil, err
	}
	return v, nil
}

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

				vars.pod, err = rk.clientset.Core().Pods("").Get(vars.pod.ObjectMeta.Name)
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
		"to_ruby": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				data, err := json.MarshalIndent(vars.pod, "", "  ")
				if err != nil {
					return nil, createException(m, err.Error())
				}

				var unmarshalled interface{}
				if err := json.Unmarshal(data, &unmarshalled); err != nil {
					return nil, createException(m, err.Error())
				}
				dumpJSON(unmarshalled, "pod")
				return nil, nil
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

type podMaker struct {
	class   *mruby.Class
	objects []podMakerInstance
}

type podMakerInstance struct {
	self *mruby.MrbValue
	vars *podMakerInstanceVars
}

type podMakerInstanceVars struct {
	pod *mruby.MrbValue
}

func newPodMakerClass(rk *RubyKube) *podMaker {
	c := &podMaker{objects: []podMakerInstance{}}
	c.class = definePodMakerClass(rk, c)
	return c
}

func definePodMakerClass(rk *RubyKube, p *podMaker) *mruby.Class {
	return defineClass(rk, "PodMaker", map[string]methodDefintion{
		"pod!": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()
				if err := standardCheck(rk, args, 1); err != nil {
					return nil, createException(m, err.Error())
				}

				container := kapi.Container{}

				// TODO: handle arrays of hashes, for multi-container pods
				params, err := hashToFlatMap(args[0], []string{"name", "image", "namespace"}, []string{"image"})
				if err != nil {
					return nil, createException(m, err.Error())
				}

				var name string

				// `hashToFlatMap` will validate that "image" key was given, so we don't need to
				// check for it; we try to split it into parts to determine the name automatically
				container.Image = params["image"]
				imageParts := strings.Split(strings.Split(container.Image, ":")[0], "/")
				name = imageParts[len(imageParts)-1]

				// if name was given, use it to override automatic name we determined from the image
				if v, ok := params["name"]; ok {
					name = v
				}

				container.Name = name
				labels := map[string]string{"name": name}

				newPodObj, err := rk.classes.Pod.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}

				pod := kapi.Pod{
					ObjectMeta: kapi.ObjectMeta{Labels: labels},
					Spec: kapi.PodSpec{
						Containers: []kapi.Container{container},
					},
				}

				if v, ok := params["namespace"]; ok {
					pod.ObjectMeta.Namespace = v
				}

				newPodObj.vars.pod = &pod

				vars.pod = newPodObj.self
				return vars.pod, nil
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

func (c *podMaker) New() (*podMakerInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := podMakerInstance{
		self: s,
		vars: &podMakerInstanceVars{nil},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *podMaker) LookupVars(this *mruby.MrbValue) (*podMakerInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *podMakerInstance) Update() (mruby.Value, error) {
	return nil, nil
}
