package rubykube

/*
  verbs.go is a collection of the verbs used to do stuff.
*/

import (
	"fmt"
	_ "io"
	_ "io/ioutil"
	_ "os"
	_ "path"
	_ "path/filepath"
	"strings"

	mruby "github.com/mitchellh/go-mruby"

	_ "k8s.io/client-go/kubernetes"
	kapi "k8s.io/client-go/pkg/api/v1"
)

// Definition is a jump table definition used for programming the DSL into the
// mruby interpreter.
type verbDefinition struct {
	verbFunc verbFunc
	argSpec  mruby.ArgSpec
}

type methodDefintion struct {
	argSpec    mruby.ArgSpec
	methodFunc mruby.Func
}

// verbJumpTable is the dispatch instructions sent to the builder at preparation time.
var verbJumpTable = map[string]verbDefinition{
	//"debug":      {debug, mruby.ArgsOpt(1)},
	//"flatten":    {flatten, mruby.ArgsNone()},
	//"tag":        {tag, mruby.ArgsReq(1)},
	//"copy":       {doCopy, mruby.ArgsReq(2)},
	//"from":       {from, mruby.ArgsReq(1)},
	//"run":        {run, mruby.ArgsAny()},
	//"user":       {user, mruby.ArgsReq(1)},
	//"with_user":  {withUser, mruby.ArgsBlock() | mruby.ArgsReq(2)},
	//"workdir":    {workdir, mruby.ArgsReq(1)},
	//"inside":     {inside, mruby.ArgsBlock() | mruby.ArgsReq(2)},
	//"env":        {env, mruby.ArgsAny()},
	//"cmd":        {cmd, mruby.ArgsAny()},
	//"entrypoint": {entrypoint, mruby.ArgsAny()},
	//"set_exec":   {setExec, mruby.ArgsReq(1)},
	"new_app":    {newApp, mruby.ArgsReq(1)},
	"count_pods": {countPods, mruby.ArgsReq(0)},
	"pods":       {pods, mruby.ArgsReq(0)},
}

var classes struct {
	Pods *podsClass
	Pod  *podClass
}

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

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

func newPodsClass(rk *RubyKube, m *mruby.Mrb) *podsClass {
	c := &podsClass{objects: []podsClassInstance{}}
	c.class = definePodsClass(rk, m, c)
	return c
}

func (c *podsClass) LookupVars(this *mruby.MrbValue) (*podsClassInstanceVars, error) {
	fmt.Println("len(c.objects) = %d\n", len(c.objects))
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func definePodsClass(rk *RubyKube, m *mruby.Mrb, p *podsClass) *mruby.Class {
	return defineClass(m, "RubyKubePods", map[string]methodDefintion{
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
		},
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
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
				if n.Fixnum()+1 > len(vars.pods.Items) {
					return nil, createException(m, "Index out of range")
				}
				pod := vars.pods.Items[n.Fixnum()]
				fmt.Printf("%d: %s/%s\n", n.Fixnum(), pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)

				newPodObj, err := classes.Pod.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				newPodObj.vars.pod = &pod
				return newPodObj.self, nil
			},
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return marshalToJSON(vars.pods, m)
			},
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

func newPodClass(rk *RubyKube, m *mruby.Mrb) *podClass {
	c := &podClass{objects: []podClassInstance{}}
	c.class = definePodClass(rk, m, c)
	return c
}

func (c *podClass) LookupVars(this *mruby.MrbValue) (*podClassInstanceVars, error) {
	for _, that := range c.objects {
		if *this == *that.self {
			return that.vars, nil
		} else {
			fmt.Printf("this(%+v) == that.self(%+v)\n", this, that.self)
		}
	}
	return nil, fmt.Errorf("could not find class instance")
}

func definePodClass(rk *RubyKube, m *mruby.Mrb, p *podClass) *mruby.Class {
	return defineClass(m, "RubyKubePod", map[string]methodDefintion{
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
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := p.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				return marshalToJSON(vars.pod, m)
			},
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

func (o *podClassInstance) Update() (mruby.Value, error) {
	v, err := o.self.Call("get!")
	if err != nil {
		return nil, err
	}
	return v, nil
}

func newApp(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if err := standardCheck(rk, args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	container := kapi.Container{}

	err := iterateRubyHash(args[0], func(key, value *mruby.MrbValue) error {
		if value.Type() != mruby.TypeString {
			return fmt.Errorf("Value for key %q is not string, must be string", key.String())
		}

		//strArgs := []string{}
		//a := value.Array()

		//for i := 0; i < a.Len(); i++ {
		//	val, err := a.Get(i)
		//	if err != nil {
		//		return err
		//	}
		//	strArgs = append(strArgs, val.String())
		//}

		switch key.String() {
		case "image":
			container.Image = value.String()
			imageParts := strings.Split(strings.Split(container.Image, ":")[0], "/")
			container.Name = imageParts[len(imageParts)-1]
		case "name":
			container.Name = value.String()
		default:
			return fmt.Errorf("new_app only accepts :image and :name as keys")
		}
		return nil
	})

	if err != nil {
		return nil, createException(m, err.Error())
	}

	pod := kapi.Pod{
		Spec: kapi.PodSpec{
			Containers: []kapi.Container{container},
		},
	}

	fmt.Printf("%#v\n", pod)

	return nil, nil
}

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	if classes.Pods == nil {
		classes.Pods = newPodsClass(rk, m)
	}

	if classes.Pod == nil {
		classes.Pod = newPodClass(rk, m)
	}

	newPodsObj, err := classes.Pods.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newPodsObj.Update(); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}

func countPods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	//if err := standardCheck(rk, args, 1); err != nil {
	//	return nil, createException(m, err.Error())
	//}

	pods, err := rk.clientset.Core().Pods("").List(kapi.ListOptions{})
	if err != nil {
		return nil, createException(m, err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	return nil, nil
}
