# `kubeplay` – a new way to interact with Kubernetes

## Usage example: easy REPL with Ruby syntax

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
kubeplay (namespace="*")> puts @pod.to_ruby # output the same as a Ruby hash
{ ... }
kubeplay (namespace="*")> @pod.delete! # I am a chaos monkey :)
```

## Resource Verbs

Currently implemented verbs are the following:

- `pods`
- `services`
- `replicasets`
- `daemonsets`

Each of these can be used with index operator, e.g. `services[10]`, as well as `first`, `last` and `any` methonds.
Any resource object can be converted to a JSON string with `to_json` method, or a Ruby hashe with `to_ruby`.


With a Ruby hash reprsentation you can do things like this:
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

You can define a verb aliases with `def_alias`, e.g. to create an `rs` verb alias for `replicasets` use
```Ruby
def_alias :rs, :replicasets
```

By default a verb operates on all namespaces, hence `(namespace="*")` is shown in the prompt.
You can switch current namespaces with `namespace` verb, e.g.
```console
kubeplay (namespace="*")> namespace "kube-system"
kubeplay (namespace="kube-system")>
```
To go back to all-namespaces mode, use `namespace "*"`.

### Resource Arguments

A verb may take up two arguments in any order - a glob string and a block or hash.


#### TL;DR

Get all replica sets in `default` namespaces which have label `app` not matching `foo` or `bar` and label `version` matching `0.1` or `0.2`:

```ruby
replicasets "default/", labels: -> { @app !~ %w(foo bar) ; @version =~ %w(0.1 0.2) ; }
```

Get all running pods with label `app` matching `foo` or `bar`:
```ruby
pods { @app =~ %w(foo bar) ; status.phase == "Running" ; }
```

#### Glob Expressions

One of the arguments would be glob expression.

Here are some examples of glob expressions.

Get all pods in `kube-systems` namespace:
```ruby
pods "kube-system/"
```

Get all pods in all namespace:
```ruby
pods "*/"
```

Get all pods in current namespace with name matching `*foo*`:
```ruby
pods "*foo*"
```

Get all pods in 

More specifically, this enables getting pods in a namespace other then current like this:
```console
kubeplay (namespace="default")> pods "kube-system/foo-*"
```
Or, gettin pods with name matching `"bar-*` in all namespace like this:
```console
kubeplay (namespace="default")> pods "*/bar-*"
```

> NOTE: if current namespace is `"*"`, `pods "*"` is the same as `pods`, and `pods "*/*"` is always the same as `pods "*/"`.

Another argument it takes is a block specifying label and field selectors using special syntax outlined below.

#### Label Selector Syntax

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
will compile a selector string `name notin (foo, bar),myapp`.

Some well-known labels have shortuct, e.g.
```ruby
{
  @app !~ %w(foo bar)
  @version =~ %w(0.1 0.2)
  @tier =~ %w(frontend backend)
}
```

Simply pass a block like this:
```ruby
replicasets { @app !~ %w(foo bar); @version =~ %w(0.1 0.2); @tier =~ %w(frontend backend); }
```

You can also use `make_label_selector` verb to construct these expressions and save those to variabels etc.

#### Field Selector Syntax

This syntax is different, yet somewhat simpler.

Here is a selector mathing all running pods:
```ruby
{ status.phase != :Running }
```

#### Using Slectors

To get all running pods with label `tier` mathcing `backend`:
```ruby
pods { status.phase != :Running ; @tier =~ "backend" ; }
```

Alternatively, if you prefer to be more explicit, you can use a hash:
```ruby
pods fields: -> { status.phase != :Running }, labels: -> { @tier =~ "backend" }
```

You can also use compose selector expressions diretly as strings, if you prefer:
```ruby
pods fields: "status.phase != Running", labels: "tier in (backend)"
```

#### Using 

## Usage example: object generator with minimal input

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
