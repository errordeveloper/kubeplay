class LabelSelector
  def initialize &block
    if block_given?
      @labels = []
      LabelSelector::Collection.new(@labels, &block)
    end
  end

  def to_s
    return @labels.reject(&:nil?).join(',')
  end

  class Key
    def initialize k, labels
      @k = k
      @labels = labels
    end

    def =~ values
      ## in Go we can simply collect in whatever we want
      @labels << LabelSelector::Expression.new(@k, :in, values)
    end

    def !~ values
      @labels << LabelSelector::Expression.new(@k, :notin, values)
    end
  end

  class Collection
    def initialize labels, &block
      @__labels__ = labels
      [:@app, :@name].each do |v|
        self.instance_variable_set(v, LabelSelector::Key.new(v.to_s.split('@')[1], labels))
      end
      self.instance_eval(&block)
    end

    def label l
      LabelSelector::Key.new(l.to_s, @__labels__)
    end
  end

  class Expression # Not needed in Go
    def initialize key, operator, values
      ## TODO go place to do validation?
      @k = key
      @o = operator
      @v = values
    end

    def to_s
      "#{@k} #{@o} (#{[@v].flatten.reject(&:nil?).join(', ')})"
    end
  end
end

class String
  def to_l # we cannot use this in Ruby, but probably can do it somehow in Go!
    LabelSelector::Key.new self
  end
end

l1 = LabelSelector.new do
  @app =~ %w(bar baz)
  label("name") =~ %w(foo foo.com example.foo.com)
  label("name") !~ %w(foo foo.com example.foo.com)
  #label("xxx") !~ %w(what.why.io)
end

puts l1

#  "xxx".to_l =~ %w(what.why.io)
#  "xxx".to_l !~ %w(what.why.io)

l2 = LabelSelector.new do
  @app !~ %w(foo)
  @name =~ %w(bar)
  @name =~ "bar"
  @name =~ 120
  @app =~ [120, :foo]
end

puts l2
