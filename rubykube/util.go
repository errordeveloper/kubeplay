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

const (
	ConverterValueTypeHash = iota
	ConverterValueTypeArray
)

type converterBranch struct {
	self     *mruby.MrbValue
	selfType int
	parent   *converterBranch
	hashKey  *mruby.MrbValue
	value    *mruby.MrbValue
	index    int
}

type Converter struct {
	branches         []*converterBranch
	branchIndex      int
	values           []*mruby.MrbValue
	isRoot           bool
	mrb              *mruby.Mrb
	unmarshalledJSON interface{}
}

func newConverter(obj interface{}, m *mruby.Mrb) (*Converter, error) {
	c := &Converter{isRoot: true, mrb: m}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &c.unmarshalledJSON); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Converter) Convert() {
	// It may be possible to rewrite this with channels and goroutines,
	// but not sure if that will be needed (and how safe it would play
	// with mruby?); Passing errors via channel might make sense...
	c.walkTree(c.unmarshalledJSON)
}

func (c *Converter) Value() mruby.Value {
	if len(c.values) > 0 {
		return c.values[0]
	}
	return nil
}

func (c *Converter) appendSimpleValue(v *mruby.MrbValue) {
	if c.isRoot {
		c.values = append(c.values, v)
		return
	}
	c.values = append(c.values, v) // for consistency
	c.thisBranch().value = v
}

func (c *Converter) appendBranch(newBranch *converterBranch) {
	if c.isRoot {
		c.isRoot = false
	} else {
		newBranch.parent = c.thisBranch()
	}
	c.values = append(c.values, newBranch.self)
	c.branches = append(c.branches, newBranch)
	i := len(c.branches) - 1
	c.branchIndex, newBranch.index = i, i
}

func (c *Converter) thisBranch() *converterBranch {
	return c.branches[c.branchIndex]
}

func (c *Converter) flipBranch() {
	if c.thisBranch().parent != nil {
		fmt.Printf("c.branchIndex = (%d => %d)\n", c.branchIndex, c.thisBranch().parent.index)
		c.branchIndex = c.thisBranch().parent.index
		c.thisBranch().value = nil
	}
}

func (c *Converter) walkTree(v interface{}) {
	switch vv := v.(type) {
	case bool:
		if vv {
			c.appendSimpleValue(c.mrb.TrueValue())
		} else {
			c.appendSimpleValue(c.mrb.FalseValue())
		}
	case string:
		c.appendSimpleValue(c.mrb.StringValue(vv))
	case float64:
		// TODO figure out if go-mruby can do it already, or it needs a patch
		// XXX should it always be a float in ruby also? (check JSON spec
		// and implementation docs for both of the parsers)
		c.appendSimpleValue(c.mrb.StringValue(fmt.Sprintf("%f", vv)))
	case map[string]interface{}:
		if len(vv) == 0 {
			fmt.Printf("(walkTree: empty map?) thisBranch[%d] = %+v\n", c.branchIndex, c.thisBranch())
			fmt.Printf("(walkTree: empty map?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
		}
		c.convertMapBranch(vv)
	case []interface{}:
		c.convertSliceBranch(vv)
	case nil:
		c.appendSimpleValue(c.mrb.NilValue())
	default:
		// XXX: should we panic here?
		fmt.Printf("(walkTree: unknown type? => %+v) thisBranch[%d] = %+v\n", vv, c.branchIndex, c.thisBranch())
	}
}

func (c *Converter) newMapBranch() *converterBranch {
	newBranch := &converterBranch{selfType: ConverterValueTypeHash}
	newHash, err := c.mrb.LoadString("{}")
	if err != nil {
		panic(fmt.Errorf("newMapBranch: %v", err))
	}
	newBranch.self = newHash
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		switch newBranch.parent.selfType {
		case ConverterValueTypeHash:
			if newBranch.parent.hashKey.String() == "securityContext" {
				fmt.Printf("(newMapBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
				fmt.Printf("(newMapBranch: reassigning?) thisBranch[%d].parent.hashKey = %+v\n", c.branchIndex, newBranch.parent.hashKey)
			}
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		case ConverterValueTypeArray: // TODO patch go-mruby to mutate arrays
			if newBranch.parent.hashKey.String() == "securityContext" {
				fmt.Printf("(newMapBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
				fmt.Printf("(newMapBranch: reassigning?) thisBranch[%d].parent.hashKey = %+v\n", c.branchIndex, newBranch.parent.hashKey)
			}
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		}
	}
	return newBranch
}

func (c *Converter) newSliceBranch() *converterBranch {
	newBranch := &converterBranch{selfType: ConverterValueTypeArray}
	newHash, err := c.mrb.LoadString("{}")
	if err != nil {
		panic(fmt.Errorf("newSliceBranch: %v", err))
	}
	newBranch.self = newHash
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		switch newBranch.parent.selfType {
		case ConverterValueTypeHash:
			if newBranch.parent.hashKey.String() == "securityContext" {
				fmt.Printf("(newSliceBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
				fmt.Printf("(newSliceBranch: reassigning?) thisBranch[%d].parent.hashKey = %+v\n", c.branchIndex, newBranch.parent.hashKey)
			}
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		case ConverterValueTypeArray:
			if newBranch.parent.hashKey.String() == "securityContext" {
				fmt.Printf("(newSliceBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
				fmt.Printf("(newSliceBranch: reassigning?) thisBranch[%d].parent.hashKey = %+v\n", c.branchIndex, newBranch.parent.hashKey)
			}
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		}
	}
	return newBranch
}

func (c *Converter) convertMapBranch(x map[string]interface{}) {
	if len(x) == 0 {
		fmt.Printf("(convertMapBranch1: empty map?) thisBranch[%d] = %+v\n", c.branchIndex, c.thisBranch())
		fmt.Printf("(convertMapBranch1: empty map?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
	}
	thisBranch := c.newMapBranch()
	if len(x) == 0 {
		fmt.Printf("(convertMapBranch2: empty map?) thisBranch[%d] = %+v\n", c.branchIndex, c.thisBranch())
		fmt.Printf("(convertMapBranch2: empty map?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
	}
	for k, v := range x {
		thisBranch.hashKey = c.mrb.StringValue(k)
		c.walkTree(v)
		if thisBranch.value != nil {
			if thisBranch.hashKey.String() == "securityContext" {
				fmt.Printf("(convertMapBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
			}
			thisBranch.self.Hash().Set(thisBranch.hashKey, thisBranch.value)
		} else {
			fmt.Printf("(nil value?) thisBranch[%d] = %+v\n", c.branchIndex, thisBranch)
		}
		if k == "securityContext" {
			fmt.Printf("k, v = %+v, %+v\n", k, v)
			fmt.Printf("thisBranch[%d] = %+v\n", c.branchIndex, thisBranch)
		}
	}
	if len(x) == 0 {
		fmt.Printf("(convertMapBranch3: empty map?) thisBranch[%d] = %+v\n", c.branchIndex, c.thisBranch())
		fmt.Printf("(convertMapBranch3: empty map?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
	}
	c.flipBranch()
}

func (c *Converter) convertSliceBranch(x []interface{}) {
	thisBranch := c.newSliceBranch()
	for k, v := range x {
		thisBranch.hashKey = c.mrb.StringValue(fmt.Sprintf("%d", k))
		c.walkTree(v)
		if thisBranch.value != nil {
			if thisBranch.hashKey.String() == "securityContext" {
				fmt.Printf("(convertSliceBranch: reassigning?) thisBranch[%d].{hashKey,value} = {%+v, %+v}\n", c.branchIndex, c.thisBranch().hashKey, c.thisBranch().value)
			}
			thisBranch.self.Hash().Set(thisBranch.hashKey, thisBranch.value)
		}
	}
	c.flipBranch()
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
