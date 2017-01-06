class LabelSelector
  def initialize *args, &block
    if block_given?
      @labels = LabelSelector::Block.new(&block)
    else args[0].is_a? Hash
      @labels = LabelSelector::Hash.new(args[0])
    end
  end

  def self.not_in *args
    [:notin] + args
  end

  def self.to_set_expr k, v
    return v if k.to_s == "___"
    if v == :present
      "#{k}"
    elsif v != :absent
      v = [v].flatten
      if v[0] == :in or v[0] == :notin
        "#{k} #{v.shift} (#{v.join(', ')})"
      else
        "#{k} in (#{v.join(', ')})"
      end
    end
  end

  def to_s
    @labels.to_s
  end

  class Block
    def initialize &block
      self.instance_eval(&block)
    end

    def not_in *args
      LabelSelector.not_in(args)
    end

    def as_hash arg
      self.instance_eval do
        @___ = [] if @___.nil?
        @___ = [@___, LabelSelector::Hash.new(arg)].flatten
      end
    end

    def method_missing m, *args
      if args.count == 0
        v = :present
      else
        v = args.flatten
      end
      self.instance_variable_set("@#{m}".to_sym, v)
    end

    def to_s
      x = self.instance_variables.map do |k|
        v = self.instance_variable_get(k)
        k = k.to_s.split('@')[1]
	LabelSelector.to_set_expr k, v
      end
      return x.reject(&:nil?).join(',')
    end
  end

  class Hash
    def initialize arg
      @hash = arg
    end

    def not_in *args
      LabelSelector.not_in(args)
    end

    def to_s
      x = @hash.map do |k,v|
        LabelSelector.to_set_expr k, v
      end
      (x + [@___].flatten).reject(&:nil?).join(',')
    end
  end
end

puts "1. instance variables and method_missing"
puts LabelSelector.new { @bar = :absent; @foo = :present; @baz = :avalue; box 2; bax(not_in(3, 4)); }

puts "2. using special ___ key to pass anything"
puts LabelSelector.new bar: :present, baz: :present, foo: "bar", ___: "food in (boxes),this!=mine"

puts "3. with as_hash"
puts LabelSelector.new { as_hash(bar: :present, baz: :present, foo: "bar", bazzz: not_in("abc")) }

puts "4. method_missing, a variable and ___"
puts LabelSelector.new { box 2; bax(not_in(3, 4)); foo; box; boxes; @foxes=not_in("boobmox"); @___="boombox!=mine"; }
puts LabelSelector.new { box 2; bax(not_in(3, 4)); foo; box; boxes; @foxes=not_in("boobmox"); as_hash(xxx: 123, ___: 456)}
puts LabelSelector.new { box 2; bax(not_in(3, 4)); foo; box; boxes; @foxes=not_in("boobmox"); @___="boombox!=mine"; as_hash(xxx: 123, ___: 456)}

puts "5. a bad example - don't do this"
puts LabelSelector.new { @foxes=not_in(as_hash({:boobmox => :mine})); @foo=as_hash(xxx: 123, ___: 456)}

## ALTERNATIVES

#pods where_lable: "foo", yields_values_in: ["bar"]
#pods where_label: "boombox", yields_values_not_in: ["mine"]
#pods [{where: "boombox", values_not_in: ["mine"]}, {where: "name"}]
#
#pods labeled: "name"
#pods labeled: "name", with_values_in: ["test"]
#pods labeled: "name", with_values_not_in: ["test"]
#
#pods "foo ~= (bar baz)"
#pods "foo ~= !(bar baz)"
#pods "foo".matches_in(["boox", "boom"])
#pods "foo".matches_not_in(["boox", "boom"])

puts "5. test"

class Label
  def initialize label_key
    @label_key = label_key
  end

  def match op, m
    "#{@label_key} #{op.to_s} (#{m})"
  end

  def join m
    ## TODO: validate
    m.map { |m| m }.join(', ')
  end

  def =~ match_set
    puts match(:in, join(match_set)) if match_set.is_a? Array
    puts match(:in, join([match_set])) if match_set.is_a? String
    puts match(:in, join([match_set.to_s])) if match_set.is_a? Symbol
    puts match(:in, join([match_set.to_s])) if match_set.is_a? Fixnum
    puts match(:in, join([match_set.to_s])) if match_set.is_a? Float
  end

  def !~ match_set
    puts match(:notin, join(match_set)) if match_set.is_a? Array
    puts match(:notin, join([match_set])) if match_set.is_a? String
    puts match(:notin, join([match_set.to_s])) if match_set.is_a? Symbol
    puts match(:notin, join([match_set.to_s])) if match_set.is_a? Fixnum
    puts match(:notin, join([match_set.to_s])) if match_set.is_a? Float
  end
end

class String
  def to_l
    Label.new self
  end
end

def label l
  Label.new l
end

l = Label.new("foo")
l =~ %w(foo foo.com example.foo.com)
l !~ %w(foo foo.com example.foo.com)

"xxx".to_l =~ %w(what.why.io)
"xxx".to_l !~ %w(what.why.io)

label "xxx" !~ %w(what.why.io)

[:@app, :@name].each { |v| instance_variable_set(v, Label.new(v.to_s.split('@')[1])) }

@app !~ %w(foo)
@name =~ %w(bar)
@name =~ "bar"
@name =~ 120
@app =~ [120, :foo]
## any undefined intance vars will be ignored 
