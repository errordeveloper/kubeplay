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
	_ "strings"

	mruby "github.com/mitchellh/go-mruby"

	_ "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
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
	"test1": {test1, mruby.ArgsReq(1)},
	"pods":  {test1, mruby.ArgsReq(0)},
}

type verbFunc func(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

func test1(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	if err := standardCheck(rk, args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	err := iterateRubyHash(args[0], func(key, value *mruby.MrbValue) error {
		if value.Type() != mruby.TypeArray {
			return fmt.Errorf("Value for key %q is not array, must be array", key.String())
		}

		strArgs := []string{}
		a := value.Array()

		for i := 0; i < a.Len(); i++ {
			val, err := a.Get(i)
			if err != nil {
				return err
			}
			strArgs = append(strArgs, val.String())
		}

		switch key.String() {
		case "a":
			fmt.Println("a:", strArgs)
		case "b":
			fmt.Println("b:", strArgs)
		default:
			return fmt.Errorf("test1 only accepts a and b as keys")
		}
		return nil
	})

	if err != nil {
		return nil, createException(m, err.Error())
	}

	return nil, nil
}

func pods(rk *RubyKube, args []*mruby.MrbValue, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	//if err := standardCheck(rk, args, 1); err != nil {
	//	return nil, createException(m, err.Error())
	//}

	pods, err := rk.clientset.Core().Pods("").List(v1.ListOptions{})
	if err != nil {
		return nil, createException(m, err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	if err != nil {
		panic(err.Error())
	}
	return nil, nil
}
