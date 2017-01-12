package rubykube

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

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

func iterateArray(arg *mruby.MrbValue, fn func(int, *mruby.MrbValue) error) error {
	array := arg.Array()
	for i := 0; i < array.Len(); i++ {
		value, err := array.Get(i)
		if err != nil {
			return err
		}

		if err := fn(i, value); err != nil {
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

type paramProcHandler func(*mruby.MrbValue) error
type params struct {
	allowed      []string
	required     []string
	skipKnown    []string
	valueType    mruby.ValueType
	procHandlers map[string]paramProcHandler
}

type paramsCollection struct {
	params    map[string]interface{}
	valueType mruby.ValueType
}

func sliceToSet(slice []string) map[string]bool {
	set := map[string]bool{}
	for _, x := range slice {
		set[x] = true
	}
	return set
}

func ValidString(value *mruby.MrbValue) *string {
	v := value.String()
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func NewParamsCollection(hash *mruby.MrbValue, spec params) (*paramsCollection, error) {
	params := make(map[string]interface{})
	validKeys := append(spec.allowed, spec.skipKnown...)
	validKeySet := sliceToSet(validKeys)
	skipKeySet := sliceToSet(spec.skipKnown)

	const (
		invalidValueTypeError = "invalid value type for %q parameter – should be %s"
		iterationError        = "failed to iterte over %s value for %q parameter – %v"
	)

	// Our params may have string, slice and map values, however there will be no nesting;
	// maps and slices are mostly to support labels and command args
	// there exist some more complex fields, e.g. volumes, but these are too complicated
	// to support with our simple generator, so we will have to provide a flattend version
	// and user will have to set things manually for the start, but eventually we can provide
	// separate helpers for volumes and other similar such stuff

	if err := iterateHash(hash, func(key0, value *mruby.MrbValue) error {
		k0 := key0.String()
		if _, ok := validKeySet[k0]; !ok {
			return fmt.Errorf("unknown parameter %q – not one of %v", k0, validKeys)
		}
		if _, ok := skipKeySet[k0]; ok {
			return nil
		}

		switch spec.valueType {
		case mruby.TypeHash:
			out := map[string]string{}
			if value.Type() != spec.valueType {
				return fmt.Errorf(invalidValueTypeError, k0, "a hash")
			}
			if err := iterateHash(value, func(key1, value *mruby.MrbValue) error {
				k1 := key1.String()
				if value.Type() != mruby.TypeString {
					return fmt.Errorf(invalidValueTypeError, fmt.Sprintf("%s::%s", k0, k1), "a string")
				}
				out[k1] = value.String()
				return nil
			}); err != nil {
				return fmt.Errorf(iterationError, "hash", k0, err)
			}
			params[k0] = out
		case mruby.TypeArray:
			out := &[]string{}
			if value.Type() != spec.valueType {
				return fmt.Errorf(invalidValueTypeError, k0, "an array")
			}
			if err := iterateArray(value, func(i1 int, value *mruby.MrbValue) error {
				if value.Type() != mruby.TypeString {
					return fmt.Errorf(invalidValueTypeError, fmt.Sprintf("%s[%d]", k0, i1), "a string")
				}
				*out = append(*out, value.String())
				return nil
			}); err != nil {
				return fmt.Errorf(iterationError, "array", k0, err)
			}
			params[k0] = *out

		default:
			if value.Type() == mruby.TypeProc {
				if fn, ok := spec.procHandlers[k0]; ok {
					if err := fn(value); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("parameter %q is a proc, but no handler found", k0)
				}
				return nil
			}
			// assume it may convert into a valid string, so it works for object that implement `to_s`
			v := ValidString(value)
			if v == nil {
				return fmt.Errorf("found invalid or empty string value for %q parameter", k0)
			}
			params[k0] = *v
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to parse given parameters – %v", err)
	}

	for _, x := range spec.required {
		if _, ok := params[x]; !ok {
			return nil, fmt.Errorf("missing required parameter %q", x)
		}
	}

	return &paramsCollection{params, spec.valueType}, nil
}

func (c *paramsCollection) ToMapOfStrings() map[string]string {
	if c.valueType != mruby.TypeString {
		return nil
	}

	out := map[string]string{}
	for k, v := range c.params {
		out[k] = v.(string)
	}
	return out
}

func (c *paramsCollection) ToMapOfMapsOfStrings() map[string]map[string]string {
	if c.valueType != mruby.TypeHash {
		return nil
	}

	out := map[string]map[string]string{}
	for k, v := range c.params {
		out[k] = v.(map[string]string)
	}
	return out
}

func (c *paramsCollection) ToMapOfSlicesOfStrings() map[string][]string {
	if c.valueType != mruby.TypeArray {
		return nil
	}

	out := map[string][]string{}
	for k, v := range c.params {
		out[k] = v.([]string)
	}
	return out
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

func getNamedMatch(re *regexp.Regexp, matchString string, submatchName string, target *string) {
	submatches := re.FindStringSubmatch(matchString)
	for i, name := range re.SubexpNames() {
		if name == submatchName {
			*target = submatches[i]
		}
	}
}

func toValues(args []*mruby.MrbValue) []mruby.Value {
	argv := []mruby.Value{}
	for _, arg := range args {
		argv = append(argv, mruby.Value(arg))
	}
	return argv
}

func call(self *mruby.MrbValue, method string, args ...*mruby.MrbValue) (mruby.Value, error) {
	v, err := self.Call(method, toValues(args)...)
	if err != nil {
		return nil, err
	}
	return v, nil
}
