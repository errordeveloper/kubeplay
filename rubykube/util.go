package rubykube

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"runtime/debug"

	mruby "github.com/mitchellh/go-mruby"
)

func init() {
	debug.SetPanicOnFault(true)
}

func createException(m *mruby.Mrb, msg string) mruby.Value {
	val, err := m.Class("Exception", nil).New(mruby.String(msg))
	if err != nil {
		panic(fmt.Sprintf("could not construct exception for return: %v", err))
	}

	return val
}

func extractStringArgs(args []*mruby.MrbValue) []string {
	strArgs := []string{}

	for _, arg := range args {
		if arg != nil && arg.Type() != mruby.TypeProc {
			strArgs = append(strArgs, arg.String())
		}
	}

	return strArgs
}

func iterateHash(arg *mruby.MrbValue, fn func(*mruby.MrbValue, *mruby.MrbValue) error) error {
	hash := arg.Hash()

	// mruby does not expose native maps, just ruby primitives, so we have to
	// iterate through it with indexing functions instead of typical idioms.
	keys, err := hash.Keys()
	if err != nil {
		return err
	}

	for i := 0; i < keys.Array().Len(); i++ {
		key, err := keys.Array().Get(i)
		if err != nil {
			return err
		}

		value, err := hash.Get(key)
		if err != nil {
			return err
		}

		if err := fn(key, value); err != nil {
			return err
		}
	}

	return nil
}

func marshalToJSON(obj interface{}, m *mruby.Mrb) (mruby.Value, mruby.Value) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, createException(m, err.Error())
	}

	return m.StringValue(string(data)), nil
}

func hashToFlatMap(hash *mruby.MrbValue, validKeys []string, requiredKeys []string) (map[string]string, error) {
	params := map[string]string{}
	validKeySet := map[string]bool{}

	const invalidTypeError = "not yet implemented – found nested %q value in %q"

	for _, x := range validKeys {
		validKeySet[x] = true
	}

	if err := iterateHash(hash, func(key, value *mruby.MrbValue) error {
		k := key.String()
		if _, ok := validKeySet[k]; !ok {
			return fmt.Errorf("unknown key %q – not one of %v", k, validKeys)
		}

		switch value.Type() {
		case mruby.TypeHash:
			return fmt.Errorf(invalidTypeError, "mruby.TypeHash", k)
		case mruby.TypeArray:
			return fmt.Errorf(invalidTypeError, "mruby.TypeArray", k)
		}

		params[k] = value.String()
		return nil
	}); err != nil {
		return nil, err
	}

	for _, x := range requiredKeys {
		if _, ok := params[x]; !ok {
			return nil, fmt.Errorf("missing required key %q", x)
		}
	}

	return params, nil
}

func checkArgs(args []*mruby.MrbValue, l int) error {
	if len(args) != l {
		return fmt.Errorf("Expected %d arg, got %d", l, len(args))
	}

	return nil
}

func standardCheck(rk *RubyKube, args []*mruby.MrbValue, l int) error {
	if err := checkArgs(args, l); err != nil {
		return err
	}

	return nil
}
