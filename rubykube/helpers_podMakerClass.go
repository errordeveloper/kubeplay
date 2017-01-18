package rubykube

import (
	"fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podMakerClassInstanceVars struct {
	pod *mruby.MrbValue
}

func newPodMakerClassInstanceVars(c *podMakerClass, s *mruby.MrbValue, args ...mruby.Value) (*podMakerClassInstanceVars, error) {
	return &podMakerClassInstanceVars{nil}, nil
}

//go:generate gotemplate "./templates/basic" "podMakerClass(\"PodMaker\", newPodMakerClassInstanceVars, podMakerClassInstanceVars)"

func (c *podMakerClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"pod!": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()
				if err := standardCheck(c.rk, args, 1); err != nil {
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

				newPodObj, err := c.rk.classes.Pod.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}

				pod := kapi.Pod{
					ObjectMeta: kapi.ObjectMeta{
						Name:   name,
						Labels: labels,
					},
					Spec: kapi.PodSpec{
						Containers: []kapi.Container{container},
					},
				}

				if v, ok := stringParams["namespace"]; ok {
					pod.ObjectMeta.Namespace = v
				}

				newPodObj.vars.pod = podTypeAlias(pod)

				vars.pod = newPodObj.self
				return vars.pod, nil
			},
			instanceMethod,
		},
	})
}

func (o *podMakerClassInstance) Update(args ...mruby.Value) (mruby.Value, error) {
	return nil, nil
}
