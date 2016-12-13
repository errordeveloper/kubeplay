App.new(service: false, image: "errordeveloper/foo:latest")
App.new(image: "errordeveloper/foo:latest")

# name label is set automaticaly, whenever possible
# otherwise user can set it with `:name`, or set all labels, or even all of metadata

Pod.new({
    image: "errordeveloper/foo:latest",
    cmd: [ ]
  },
  {
    image: "errordeveloper/foo:latest",
    cmd: [ ]
  })

c1 = Container.new(image: "errordeveloper/foo:latest")
c2 = c1.clone
c2.image = "errordeveloper/bar:latest"
c.cmd = [ ]

# container names are infered from image name, whenever possible, otherwise an exception is thrown
# user can set a container name

...

# declarative syntax with Kubefile

image "errordeveloper/foo:latest"
replicas "10"
labels {
  myapp: "foo"
}

# those are all Ruby methods that can only be called once per Kubefile, they result in a replicaset/service pair (unless
# specified otherwise) and generate data that is passed to Go implementation further validation (binding to Go may be
# seen as a premature optimisation, but something to consider also)
