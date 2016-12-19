package rubykube

/*
  funcs.go provides simple functions for use within rk that do *not*
  They are intended to be used as gathering functions for predicates and templating.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	_ "strings"

	mruby "github.com/mitchellh/go-mruby"
)

type funcDefinition struct {
	fun     funcFunc
	argSpec mruby.ArgSpec
}

type funcFunc func(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value)

// mrubyJumpTable is the dispatch instructions sent to the mruby interpreter at rk setup.
var funcJumpTable = map[string]funcDefinition{
	"import": {importFunc, mruby.ArgsReq(1)},
	"getenv": {getenv, mruby.ArgsReq(1)},
}

// importFunc implements the import function.
//
// import loads a new ruby file at the point of the function call. it is
// principally used to extend and consolidate reusable code.
func importFunc(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	args := m.GetArgs()
	if err := checkArgs(args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	content, err := ioutil.ReadFile(args[0].String())
	if err != nil {
		return nil, createException(m, err.Error())
	}

	val, err := rk.Run(string(content))
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return val, nil
}

// getenv retrieves a value from the environment (passed in as string)
// and returns a string with the value. If no value exists, an empty string is
// returned.
func getenv(rk *RubyKube, m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
	args := m.GetArgs()

	if err := standardCheck(rk, args, 1); err != nil {
		return nil, createException(m, err.Error())
	}

	return mruby.String(os.Getenv(args[0].String())), nil
}

const (
	classMethod = iota
	instanceMethod
)

type methodDefintion struct {
	argSpec    mruby.ArgSpec
	methodFunc mruby.Func
	methodType int
}

func defineClass(m *mruby.Mrb, name string, methods map[string]methodDefintion) *mruby.Class {
	class := m.DefineClass(name, nil)
	for name, m := range methods {
		if m.methodType == classMethod {
			class.DefineClassMethod(name, m.methodFunc, m.argSpec)
		} else {
			class.DefineMethod(name, m.methodFunc, m.argSpec)
		}
	}
	return class
}

func bytesToStringSlice(data []byte, chunkSize int) (chunks []string) {
	stringData := string(data)
	for {
		if len(stringData) >= chunkSize {
			chunks, stringData = append(chunks, stringData[0:chunkSize]), stringData[chunkSize:]
		} else {
			chunks, stringData = append(chunks, stringData[0:len(stringData)]), ""
			break
		}
	}
	return chunks
}

func marshalToJSON(obj interface{}, m *mruby.Mrb) (mruby.Value, mruby.Value) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, createException(m, err.Error())
	}

	value, err := m.LoadString(fmt.Sprintf("@tmp = []"))
	if err != nil {
		return nil, createException(m, err.Error())
	}

	// MRB_PARSER_BUF_SIZE is 1024, but we get extra escaping characters,
	// so 1024/4 seems reasonably safe...
	for _, chunk := range bytesToStringSlice(data, 1024/4) {
		_, err := m.LoadString(fmt.Sprintf("@tmp << %q", chunk))
		if err != nil {
			return nil, createException(m, err.Error())
		}
	}

	return value, nil
}

func intToValue(m *mruby.Mrb, i int) (mruby.Value, mruby.Value) {
	value, err := m.LoadString(fmt.Sprintf("%d", i))
	if err != nil {
		return nil, createException(m, err.Error())
	}
	return value, nil
}
