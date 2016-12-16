package rubykube

/*
  verbs.go is a collection of the verbs used to do stuff.
*/

import (
	"encoding/json"
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

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

type podClass struct {
	class *mruby.Class
	self  *mruby.MrbValue
	vars  podInstanceVars
}

type podInstanceVars struct {
	pod *kapi.Pod
}

type podsClass struct {
	class *mruby.Class
	self  *mruby.MrbValue
	vars  podsInstanceVars
}
type podsInstanceVars struct {
	pods *kapi.PodList
}

var (
	// TODO: this is obviously dangerous, the Pod and PodList objects should be instance variables,
	// but it may be sufficent for the moment – as the ways you can instantiate classes is limited
	// (or at least it's supposed to be limited)
	podObj  podClass
	podsObj podsClass
)

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

func defineClass(m *mruby.Mrb, name string, methods map[string]methodDefintion) *mruby.Class {
	class := m.DefineClass(name, nil)
	for name, m := range methods {
		class.DefineMethod(name, m.methodFunc, m.argSpec)
	}
	return class
}

func marshalToJSON(obj interface{}, m *mruby.Mrb) (mruby.Value, mruby.Value) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, createException(m, err.Error())
	}
	fmt.Printf("%s", data)
	// TODO: should really return a ruby string
	return nil, nil
}

func definePodsClass(rk *RubyKube, m *mruby.Mrb) *mruby.Class {
	return defineClass(m, "RubyKubePods", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				var err error

				podsObj.vars.pods, err = rk.clientset.Core().Pods("").List(kapi.ListOptions{})
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return self, nil
			},
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				pods := podsObj.vars.pods

				for n, pod := range pods.Items {
					fmt.Printf("%d: %s/%s\n", n, pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
				}
				return self, nil
			},
		},
		"[]": {
			mruby.ArgsReq(1), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				var err error
				pods := podsObj.vars.pods

				args := m.GetArgs()
				err = standardCheck(rk, args, 1)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				n := args[0]
				if n.Type() != mruby.TypeFixnum {
					return nil, createException(m, "Argument must be a integer")
				}
				if n.Fixnum()+1 > len(pods.Items) {
					return nil, createException(m, "Index out of range")
				}
				pod := pods.Items[n.Fixnum()]
				fmt.Printf("%d: %s/%s\n", n.Fixnum(), pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)

				podObj.self, err = podObj.class.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				podObj.vars.pod = &pod
				return podObj.self, nil
			},
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				pods := podsObj.vars.pods
				return marshalToJSON(pods, m)
			},
		},
	})
}

func definePodClass(rk *RubyKube, m *mruby.Mrb) *mruby.Class {
	return defineClass(m, "RubyKubePod", map[string]methodDefintion{
		"get!": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				var err error
				pod := podObj.vars.pod

				pod, err = rk.clientset.Core().Pods("").Get(pod.ObjectMeta.Name)
				if err != nil {
					return nil, createException(m, err.Error())
				}
				return self, nil
			},
		},
		"inspect": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				pod := podObj.vars.pod
				fmt.Printf("self: %s/%s\n", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
				return self, nil
			},
		},
		"to_json": {
			mruby.ArgsReq(0), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				pod := podObj.vars.pod
				return marshalToJSON(pod, m)
			},
		},
	})
}

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value *mruby.MrbValue
		err   error
	)

	if podObj.class == nil {
		podObj = podClass{
			class: definePodClass(rk, m),
			vars:  podInstanceVars{&kapi.Pod{}},
		}
	}
	if podsObj.class == nil {
		podsObj = podsClass{
			class: definePodsClass(rk, m),
			vars:  podsInstanceVars{&kapi.PodList{}},
		}
	}

	if value, err = podsObj.class.New(); err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = value.Call("get!"); err != nil {
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
