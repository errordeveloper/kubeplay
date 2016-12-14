package rubykube

import (
	_ "fmt"
	"strings"

	"github.com/erikh/box/builder/signal"
	"github.com/erikh/box/log"
	mruby "github.com/mitchellh/go-mruby"
)

type RubyKube struct {
	mrb *mruby.Mrb
}

func keep(omitFuncs []string, name string) bool {
	for _, fun := range omitFuncs {
		if name == fun {
			return false
		}
	}
	return true
}

// NewRubyKube may return an error on mruby issues.
func NewRubyKube(omitFuncs []string) (*RubyKube, error) {
	rk := &RubyKube{mrb: mruby.NewMrb()}

	rk.mrb.DisableGC()

	for name, def := range verbJumpTable {
		if keep(omitFuncs, name) {
			rk.AddVerb(name, def.verbFunc, def.argSpec)
		}
	}

	for name, def := range funcJumpTable {
		if keep(omitFuncs, name) {
			inner := def.fun
			fn := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				return inner(rk, m, self)
			}

			rk.mrb.TopSelf().SingletonClass().DefineMethod(name, fn, def.argSpec)
		}
	}

	signal.SetSignal(nil)

	return rk, nil
}

// AddVerb adds a function to the mruby dispatch as well as adding hooks around
// the call to ensure containers are committed and intermediate layers are
// cleared.
func (rk *RubyKube) AddVerb(name string, fn verbFunc, args mruby.ArgSpec) {
	hookFunc := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
		args := m.GetArgs()
		strArgs := extractStringArgs(args)

		log.BuildStep(name, strings.Join(strArgs, ", "))

		return fn(rk, args, m, self)
	}

	rk.mrb.TopSelf().SingletonClass().DefineMethod(name, hookFunc, args)
}

// Run the script.
func (rk *RubyKube) Run(script string) (*mruby.MrbValue, error) {
	if _, err := rk.mrb.LoadString(script); err != nil {
		return nil, err
	}

	return mruby.String("").MrbValue(rk.mrb), nil
}

// Close tears down all functions of the RubyKube, preparing it for exit.
func (rk *RubyKube) Close() error {
	rk.mrb.EnableGC()
	rk.mrb.FullGC()
	rk.mrb.Close()
	return nil
}
