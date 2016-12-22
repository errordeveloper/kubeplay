package converter

import (
	"encoding/json"
	"fmt"

	mruby "github.com/mitchellh/go-mruby"
)

const (
	valueTypeHash = iota
	valueTypeArray
)

type converterBranch struct {
	self     *mruby.MrbValue
	selfType int              // this to say whether we are working with a hash or an array
	parent   *converterBranch // pointer to the parent, so we know where to go back once done
	hashKey  *mruby.MrbValue  // hash key to use for the current value (if it is in a hash)
	value    *mruby.MrbValue  // current value we are handling
	index    int              // our index in the tree
}

type Converter struct {
	branches    []*converterBranch // list of tree branches – hashes and arrays
	branchIndex int                // maps to the current position in the tree
	values      []*mruby.MrbValue  // used to store all values we find
	isRoot      bool               // idiates we are at the root of the tree
	mrb         *mruby.Mrb         // local instance of mruby
}

// New returns a new converter for any Kubernetes API object to Ruby
func New(m *mruby.Mrb) *Converter {
	return &Converter{isRoot: true, mrb: m}
}

// Convert performs conversion from any Kubernetes API object to Ruby,
// it may panic, if there is something terribly wrong; it's effective
// in wire-format and uses lower-case keys, unlike Go structs.
func (c *Converter) Convert(obj interface{}) error {
	// It may be possible to rewrite this with channels and goroutines,
	// but not sure if that will be needed (and how safe it would play
	// with mruby?); Passing errors via channel might make sense...
	if len(c.values) > 0 || len(c.branches) > 0 || !c.isRoot {
		return fmt.Errorf("Convert: don't call me again, I'm stupid!")
	}

	// As user is expected to be falimial with wire format of the API,
	// we convert it to JSON first. Also, it'd require a code generator
	// to provide conversion if we didn't do this.
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// Re-encode JSON data into an interface and store it
	var tree interface{}
	if err := json.Unmarshal(data, &tree); err != nil {
		return err
	}

	c.walkTree(tree)
	return nil
}

// Value returns converted Ruby value (currently a hash)
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
	c.values = append(c.values, v) // keep this for consistency
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
		// fmt.Printf("c.branchIndex = (%d => %d)\n", c.branchIndex, c.thisBranch().parent.index)
		c.branchIndex = c.thisBranch().parent.index
		c.thisBranch().value = nil // reset the value, so we don't inherit it unitentionally
	}
}

func (c *Converter) walkTree(v interface{}) {
	// we enter an interface and look at its type
	switch vv := v.(type) {
	// it may be a simple one (e.g. bool, string or float64),
	// in which case we simple pick up the value and append it
	// to the internal list of values
	// this code is generic enough to treat simple non-nested values
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
	// the root is going to be an map, but we also handle slices
	// in both cases we recurse into the data and collect all what's there
	case map[string]interface{}:
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
	newBranch := &converterBranch{selfType: valueTypeHash}
	newHash, err := c.mrb.LoadString("{}")
	if err != nil {
		panic(fmt.Errorf("newMapBranch: %v", err))
	}
	newBranch.self = newHash
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		switch newBranch.parent.selfType {
		case valueTypeHash:
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		case valueTypeArray: // TODO patch go-mruby to mutate arrays
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		}
	}
	return newBranch
}

func (c *Converter) convertMapBranch(x map[string]interface{}) {
	thisBranch := c.newMapBranch()
	for k, v := range x {
		thisBranch.hashKey = c.mrb.StringValue(k)
		c.walkTree(v)
		if thisBranch.value != nil {
			thisBranch.self.Hash().Set(thisBranch.hashKey, thisBranch.value)
		}
	}
	c.flipBranch()
}

func (c *Converter) newSliceBranch() *converterBranch {
	newBranch := &converterBranch{selfType: valueTypeArray}
	newHash, err := c.mrb.LoadString("{}")
	if err != nil {
		panic(fmt.Errorf("newSliceBranch: %v", err))
	}
	newBranch.self = newHash
	c.appendBranch(newBranch)
	if newBranch.parent != nil {
		switch newBranch.parent.selfType {
		case valueTypeHash:
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		case valueTypeArray:
			newBranch.parent.self.Hash().Set(newBranch.parent.hashKey, newHash)
		}
	}
	return newBranch
}

func (c *Converter) convertSliceBranch(x []interface{}) {
	thisBranch := c.newSliceBranch()
	for k, v := range x {
		thisBranch.hashKey = c.mrb.StringValue(fmt.Sprintf("%d", k))
		c.walkTree(v)
		if thisBranch.value != nil {
			thisBranch.self.Hash().Set(thisBranch.hashKey, thisBranch.value)
		}
	}
	c.flipBranch()
}
