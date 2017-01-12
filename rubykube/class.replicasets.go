package rubykube

import (
	"fmt"
	_ "math/rand"

	_ "github.com/errordeveloper/kubeplay/rubykube/converter"
	mruby "github.com/mitchellh/go-mruby"
	//kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type replicaSetsClass struct {
	class   *mruby.Class
	objects []replicaSetsClassInstance
	rk      *RubyKube
}

type replicaSetsClassInstance struct {
	self *mruby.MrbValue
	vars *replicaSetsClassInstanceVars
}

type replicaSetsClassInstanceVars struct {
	replicaSets *kext.ReplicaSetList
}

func newReplicaSetsClass(rk *RubyKube) *replicaSetsClass {
	c := &replicaSetsClass{objects: []replicaSetsClassInstance{}, rk: rk}
	c.class = defineReplicaSetsClass(rk, c)
	return c
}

func defineReplicaSetsClass(rk *RubyKube, r *replicaSetsClass) *mruby.Class {
	return defineClass(rk, "RepicaSets", map[string]methodDefintion{
		"object_count": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return m.FixnumValue(len(r.objects)), nil
			},
			classMethod,
		},
	})
}

func (c *replicaSetsClass) New() (*replicaSetsClassInstance, error) {
	s, err := c.class.New()
	if err != nil {
		return nil, err
	}
	o := replicaSetsClassInstance{
		self: s,
		vars: &replicaSetsClassInstanceVars{
			&kext.ReplicaSetList{},
		},
	}
	c.objects = append(c.objects, o)
	return &o, nil
}

func (c *replicaSetsClass) LookupVars(this *mruby.MrbValue) (*replicaSetsClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func (o *replicaSetsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
