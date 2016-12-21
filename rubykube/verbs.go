package rubykube

/*
  verbs.go is a collection of the verbs used to do stuff.
*/

import (
	_ "fmt"
	"strings"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

// Definition is a jump table definition used for programming the DSL into the
// mruby interpreter.
type verbDefinition struct {
	verbFunc verbFunc
	argSpec  mruby.ArgSpec
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
	"newPod": {newPodFromImage, mruby.ArgsReq(1)},
	"pods":   {pods, mruby.ArgsReq(0)},
}

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

func newPodFromImage(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if err := standardCheck(rk, args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	container := kapi.Container{}

	params, err := hashToFlatMap(args[0], []string{"name", "image"}, []string{"image"})
	if err != nil {
		return nil, createException(m, err.Error())
	}

	// `hashToFlatMap` will validate that "image" key was given, so we don't need to
	// check for it; we try to split it into parts to determine the name automatically
	container.Image = params["image"]
	imageParts := strings.Split(strings.Split(container.Image, ":")[0], "/")
	container.Name = imageParts[len(imageParts)-1]

	// if name was given, use it to override automatic name we determined from the image
	if name, ok := params["name"]; ok {
		container.Name = name
	}

	pod := kapi.Pod{
		Spec: kapi.PodSpec{
			Containers: []kapi.Container{container},
		},
	}

	return marshalToJSON(pod, m)
}

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	var (
		value mruby.Value
		err   error
	)

	newPodsObj, err := rk.classes.Pods.New()
	if err != nil {
		return nil, createException(m, err.Error())
	}

	if value, err = newPodsObj.Update(); err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}
