package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podMakerClass struct {
	class   *mruby.Class
	objects []podMakerInstance
	rk      *RubyKube
}

type podMakerInstance struct {
	self *mruby.MrbValue
	vars *podMakerInstanceVars
}

type podMakerInstanceVars struct {
	pod *mruby.MrbValue
}

func newPodMakerClass(rk *RubyKube) *podMakerClass {
	c := &podMakerClass{objects: []podMakerInstance{}, rk: rk}
	c.class = definePodMakerClass(rk, c)
	return c
}

func definePodMakerClass(rk *RubyKube, p *podMakerClass) *mruby.Class {
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

				if args[0].Type() != mruby.TypeHash {
					return nil, createException(m, "First argument must be a hash")
				}

				// TODO: handle arrays of hashes, for multi-container pods
				stringParamsCol, err := NewParamsCollection(args[0],
					params{
						allowed:   []string{"image", "name", "namespace"},
						required:  []string{"image"},
						skipKnown: []string{"labels", "env", "command"},
						valueType: mruby.TypeString,
					},
				)

				if err != nil {
					return nil, createException(m, err.Error())
				}

				stringParams := stringParamsCol.ToMapOfStrings()

				hashParamsCol, err := NewParamsCollection(args[0],
					params{
						allowed:   []string{"labels", "env"},
						required:  []string{},
						skipKnown: []string{"image", "name", "namespace", "command"},
						valueType: mruby.TypeHash,
					},
				)

				if err != nil {
					return nil, createException(m, err.Error())
				}

				arrayParamsCol, err := NewParamsCollection(args[0],
					params{
						allowed:   []string{"command"},
						required:  []string{},
						skipKnown: []string{"image", "name", "namespace", "labels", "env"},
						valueType: mruby.TypeArray,
					},
				)

				if err != nil {
					return nil, createException(m, err.Error())
				}

				fmt.Printf("stringParams=%+v\nhashParams=%+v\narrayParams=%+v\n", stringParams, hashParamsCol.ToMapOfMapsOfStrings(), arrayParamsCol.ToMapOfSlicesOfStrings())

				container := kapi.Container{}
				var name string

				// `hashArgsToSimpleMap` will validate that "image" key was given, so we don't need to
				// check for it; we try to split it into parts to determine the name automatically
				container.Image = stringParams["image"]
				imageParts := strings.Split(strings.Split(container.Image, ":")[0], "/")
				name = imageParts[len(imageParts)-1]

				// if name was given, use it to override automatic name we determined from the image
				if v, ok := stringParams["name"]; ok {
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

				if v, ok := stringParams["namespace"]; ok {
					pod.ObjectMeta.Namespace = v
				}

				p := podTypeAlias(pod)
				newPodObj.vars.pod = &p

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

func (c *podMakerClass) New() (*podMakerInstance, error) {
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

func (c *podMakerClass) LookupVars(this *mruby.MrbValue) (*podMakerInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *podMakerInstance) Update(args ...mruby.Value) (mruby.Value, error) {
	return nil, nil
}
