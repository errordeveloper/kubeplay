module KubeShellObjectPods
  def [](args)
    puts "kubectl get pods", args
  end
  def to_s
    "pods as a string"
  end
  def inspect
    "pods are us!"
  end
end

module KubeShellObject
  def [](args)
    puts args
  end
end

def _pods
  @this = Object.new
  @this.extend KubeShellObjectPods
  return @this
end

def _
  @this = Object.new
  @this.extend KubeShellObject
  return @this
end

_[name: "test-label", tada: "yes", foo: "bar", bar: true]

_pods[foo: 1]
puts _pods
