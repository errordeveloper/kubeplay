# `kubeplay` – a new way to interact with Kubernetes

## Usage: easy REPL with Ruby syntax

```console
> ./kubeplay
kubeplay (namespace="*")> pods # list pods in the cluster
<list-of-pods>
kubeplay (namespace="*")> @pod = _.any # pick a random pod from the list
kubeplay (namespace="*")> puts @pod.to_json # output the pod definition in JSON
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
kubeplay> (namespace="*") puts @pod.to_ruby # output the same as a Ruby hash
{ ... }
kubeplay (namespace="*")> @pod.delete!
kubeplay> ^D
> 
```

All resources can be converted to a Ruby-native reprsentation, which means you can do things like this:
```ruby
@metadata = replicasets("*/").to_ruby.items.map do |k,v|
   v.metadata
end

@metadata.each do |i|
    puts "Name:\t#{i.name}"
    puts "Labels:\t#{i.labels}"
    puts
end
```

By default commands operate on all namespaces, hence `(namespace="*")` is shown in the prompt.
You can switch current namespaces with `namespace` command, e.g.
```console
kubeplay (namespace="*")> namespace "kube-system"
kubeplay (namespace="kube-system")>
```
To go back to all-namespaces mode, use `namespace "*"`.

The `pods` verb may take up two arguments in any order, a glob expressions, e.g.
```console
kubeplay (namespace="*")> pods "kube-system/"
```
where you can use `"*/"` to look at pods in all namespace.

For example, you can get all pods in a namespace other then current like this:
```console
kubeplay (namespace="default")> pods "kube-system/"
```
Or, you can get pods in all namespace like this:
```console
kubeplay (namespace="default")> pods "*/"
```

To get all pods containing `foo` in current namespace:
```console
kubeplay (namespace="default")> pods "*foo*"
```
To get all pods begining with `bar-*` in all namespaces:

```console
kubeplay (namespace="default")> pods "*/bar-*"

```

> NOTE: `pods "*"` is the same as `pods`, and `pods "*/*"` is the same as `pods "*/"`.

Another arument it takes is a hash, where keys `:labels` and `:fields` are recognised.

The `:labels` can be passed in as a string or as a block with special syntax described bellow.

To match a label agains a set of values, use `label("name") =~ %w(foo bar)`, or `!~`.

If you want to just get resources with a certain label to set anything, use `label("baz").defined?`

This
```ruby
{
  label("name") =~ %w(foo bar)
  label("baz").defined?
}
```
will compile a selector string `name in (foo, bar),myapp`.

And this
```ruby
{
  label("name") !~ %w(foo bar)
  label("baz").defined?
}
```
will compile a selector string `name noin (foo, bar),myapp`.

Some well-known labels have shortuct, e.g.
```ruby
{
  @app !~ %w(foo bar)
  @version =~ %w(0.1 0.2)
  @tier =~ %w(frontend backend)
}
```

You can use `make_label_selector` verb to construct these expressions, or simply pass a lambda like this:
```ruby
replicasets labels: -> { @app !~ %w(foo bar); @version =~ %w(0.1 0.2); @tier =~ %w(frontend backend); }
```

Another allowed key for the hash argument of `pods` verb is `:fields`, which can be used to match resource fields.
Currently this doesn't have special syntax and a string must be constructed, e.g.
```ruby
pods fields: "status.phase!=Running", labels: -> { @tier =~ "backend" }
```

## Usage: object generator with minimal input

```console
> ./kubeplay -kubeconfig ~/.kube/config
kubeplay (namespace="*")> make_pod(image: "errordeveloper/foo:latest").to_json
kubeplay (namespace="*")> puts _
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
kubeplay (namespace="*")> ^D
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
