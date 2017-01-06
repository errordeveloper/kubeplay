package rubykube

var patches = []string{
	// We need to allow `Struct`-alike access to attributes `Struct` doesn't just work, and we need to do a lot
	// of recursion etc, and it gets complicated... There exists `OpenStruct` and `RecursiveOpenStruct`, but
	// both depend on `Thread` and `Regexp`. There is a version of `OpenStruct` ported to mruby, but it's
	// quite tedious to make it recursive... Monkey-patching `Hash` (like in http://stackoverflow.com/a/38175437/717998)
	// works well for our simple DSL case, at least for now.
	// We can do something smarter later, but this just does what we want. The only thing we will need to ensure
	// is to not try to define methods that already map to attributes, or at least be careful when we do that.
	// TODO: how should we handle setters? It seems a big question in general... Pods are immutable, but we should
	// let user update a value of some filed and apply it, e.g.
	// `my = rc.any; my.rc.spec.template.containers << myHandyDebuggerSideCar; myrc.apply!`
	`class Hash
	  def method_missing(m, *opts)
	    if self.has_key?(m.to_s)
	      return self[m.to_s]
	    elsif self.has_key?(m.to_sym)
	      return self[m.to_sym]
	    end
	    return nil
	  end
	 end
	`,
	// This is needed for creating lable expressions
	`class String
	   def to_l
	     RubyKube::LabelName.new self
	   end
	 end
	`,
}

func (rk *RubyKube) applyPatches() error {
	for _, p := range patches {
		if _, err := rk.mrb.LoadString(p); err != nil {
			return err
		}
	}
	return nil
}
