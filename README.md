# `kubeshell` – a new way to interact with Kubernetes

## Usage: easy REPL with Ruby syntax

```console
> ./kubeshell -kubeconfig ~/.kube/config
kubeshell> pods # list pods in the cluster
<list-of-pods>
kubeshell> @pod = _.any # pick a random pod from the list
kubeshell> puts @pod.to_json # output the pod definition in JSON
{
  "metadata": {
    ...
  },
  "spec": {
    ...
    "containers": [
      {
        ...
      }
    ],
  },
  "status": {
    ...
  }
}
kubeshell> puts @pod.to_ruby # output the same as a Ruby hash
{ ... }
kubeshell> @pod.delete!
kubeshell> ^D
> 
```

## Usage: object generator with minimal input

```console
> ./kubeshell -kubeconfig ~/.kube/config
kubeshell> new.pod!(image: "errordeveloper/foo:latest").to_json
+++ Execute: new 
kubeshell> puts _
{
  "metadata": {
    "creationTimestamp": null,
    "labels": {
      "name": "foo"
    }
  },
  "spec": {
    "containers": [
      {
        "name": "foo",
        "image": "errordeveloper/foo:latest",
        "resources": {}
      }
    ]
  },
  "status": {}
}
kubeshell> ^D
>
```

### Building

Get the source code and build the dependencies:

```bash
go get github.com/Masterminds/glide
go get -d github.com/errordeveloper/kubeshell
cd $GOPATH/src/github.com/errordeveloper/kubeshell
$GOPATH/bin/glide up
make -C vendor/github.com/mitchellh/go-mruby libmruby.a
go install ./rubykube
```

Build `kubeshell`:
```bash
go build .
```

### Credits

The mruby integration was inspired by [@erikh's box](https://github.com/erikh/box), and some of the code was initially copied from there.
