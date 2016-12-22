# `kubeshell` – a new way to interact with Kubernetes API from the terminal

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
