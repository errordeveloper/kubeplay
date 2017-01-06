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
      self.instance_eval { @___ = LabelSelector::Hash.new arg }
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
      x + @___.to_s unless @___.nil?
      return x.reject(&:nil?).join(',')
    end
  end
end

puts LabelSelector.new { @bar = :absent; @foo = :present; @baz = :avalue; box 2; bax(not_in(3, 4)); }

puts LabelSelector.new bar: :present, baz: :present, foo: "bar", ___: "food in (boxes),this!=mine"

puts "3. with as_hash"
puts LabelSelector.new { as_hash(bar: :present, baz: :present, foo: "bar", bazzz: not_in("abc")) }

puts "4. using mostly method_missing"
puts LabelSelector.new { box 2; bax(not_in(3, 4)); foo; box; boxes; @foxes=not_in("boobmox"); }
