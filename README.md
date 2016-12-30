# `kubeplay` – a new way to interact with Kubernetes

## Usage: easy REPL with Ruby syntax

```console
> ./kubeplay -kubeconfig ~/.kube/config
kubeplay> pods # list pods in the cluster
<list-of-pods>
kubeplay> @pod = _.any # pick a random pod from the list
kubeplay> puts @pod.to_json # output the pod definition in JSON
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
kubeplay> puts @pod.to_ruby # output the same as a Ruby hash
{ ... }
kubeplay> @pod.delete!
kubeplay> ^D
> 
```

## Usage: object generator with minimal input

```console
> ./kubeplay -kubeconfig ~/.kube/config
kubeplay> new.pod!(image: "errordeveloper/foo:latest").to_json
+++ Execute: new 
kubeplay> puts _
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
kubeplay> ^D
>
```

### Building

Get the source code and build the dependencies:

```bash
go get github.com/Masterminds/glide
go get -d github.com/errordeveloper/kubeplay
cd $GOPATH/src/github.com/errordeveloper/kubeplay
$GOPATH/bin/glide up
make -C vendor/github.com/mitchellh/go-mruby libmruby.a
go install ./rubykube
```

Build `kubeplay`:
```bash
go build .
```

### Credits

The mruby integration was inspired by [@erikh's box](https://github.com/erikh/box), and some of the code was initially copied from there.
