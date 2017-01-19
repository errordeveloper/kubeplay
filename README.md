# `kubeplay` – a new way to interact with Kubernetes

> NOTE: _this project is still in an early stage_

> If you like this project, please checkout [TODOs](#todos) and open an issue if you'd like to contribute or discuss anything in particular.

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
Any resource object can be converted to a JSON string with `to_json` method, or a Ruby object with `to_ruby`.

With a Ruby object reprsentation you can do things like this:
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

To get all replica sets in `default` namespaces which have label `app` not matching `foo` or `bar` and label `version` matching `0.1` or `0.2` use

```ruby
replicasets "default/", labels: -> { @app !~ %w(foo bar) ; @version =~ %w(0.1 0.2) ; }
```

To get all running pods with label `app` matching `foo` or `bar` use
```ruby
pods { @app =~ %w(foo bar) ; status.phase == "Running" ; }
```

#### Glob Expressions

Here are some examples illustrating the types of glob expressions that `kubeplay` understands.

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

More specifically, this enables getting pods in a namespace other then current like this:
```console
kubeplay (namespace="default")> pods "kube-system/foo-*"
```
Or, gettin pods with name matching `"bar-*` in all namespace like this:
```console
kubeplay (namespace="default")> pods "*/bar-*"
```

> NOTE: if current namespace is `"*"`, `pods "*"` is the same as `pods`; `pods "*/*"` is always the same as `pods "*/"`.

#### Label & Field Selectors

Another argument a resource verb understand is a block specifying label and field selectors using special syntax outlined below.

##### Label Selector Syntax

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

##### Field Selector Syntax

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

### Inspecting the Logs

To get grep logs for any pod matching given selector

```ruby
pods{ @name =~ "launch-generator" ; }.any.logs.grep ".*INFO:.*", ".*user-agent:.*"
```

## Usage example: object generator with minimal input

```console
> ./kubeplay -kubeconfig ~/.kube/config
kubeplay (namespace="*")> @pod = make_pod(image: "errordeveloper/foo:latest")
kubeplay (namespace="*")> puts _.to_json
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
kubeplay (namespace="*")> @pod.create!
kubeplay (namespace="*")> @pod.delete!

kubeplay (namespace="*")> ^D
>
```

### TODOs

- [x] `pod.delete!`
- [x] `pod.create!`
- [x] `pod.logs` & `pod.logs.grep`
- [x] `pods.logs` & `pods.logs.grep`
- [ ] `pod.logs.pager` and `pod.logs.grep.pager`
- [ ] grep logs in any set of resources
- [ ] more fluent behaviour of set resources, e.g. `replicasets.pods` and not `replicasets.any.pods`
- [ ] reverse lookup, e.g. given `@rs = replicasets.any`, `@rs.pods.any.owner` should be the same as `@rs`
- [ ] way to run scripts and not just REPL
- [ ] extend resource generator functionality
  - [ ] `ReplicaSet`+`Service`
  - [ ] `Kubefile` DSL
- [ ] other ideas
  - [ ] simple controller loop framework
  - [ ] multi-cluster support
  - [ ] resource diff
  - [ ] network policy tester framework
  - [ ] eval/exec code in a pod
  - [ ] test framework for apps, e.g. "Here is my app, it has a configmap and a secrete, and I want to test if it works"

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
