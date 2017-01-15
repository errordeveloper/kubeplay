class Fields
  def initialize parent
    @path = [parent].flatten.reject(&:nil?)
  end

  def method_missing name
    Fields.new @path << name
  end

  def to_s
    @path.join('.')
  end

  def == value
    "#{self}==#{value}"
  end

  def != value
    "#{self}!=#{value}"
  end
end

f = Fields.new nil
puts f.bar.to_s
puts f.bar.baz.to_s
puts f.bar.baz.bar.to_s

puts f.bar.not.baz != :Ready
puts f.bar.not.baz == :Ready
