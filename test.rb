module KubeShellObjectPods
  def [](args)
    puts "kubectl get pods", args
  end
end

module KubeShellObject
  def [](args)
    puts args
  end
end

pods = Object.new
pods.extend KubeShellObjectPods

_ = Object.new
_.extend KubeShellObject

_[name: "test-label", tada: "yes", foo: "bar", bar: true]

pods[foo: 1]
