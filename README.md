# allar/kitgen
Bad experiment in learning go-kit.

I'm very new to Golang and I'm having trouble with the amount of code that is required for go-kit.

The goal of this is to have a go-kit generator that has a simple command structure and allow for entire service and project regeneration with as little thought of possible.

# How

The vast majority of this code and generated code comes from work by Kujtim Hoxha (https://github.com/kujtimiihoxha), who created the GoKit Cli (https://github.com/kujtimiihoxha/kit) as well as a series of repos located at https://github.com/go-services.

Very, very little of this is my own.

# Unsupported

While I aim to update this as I learn more and develop my own microservices with go-kit, I don't have any formal plans for supporting this. It could already be dead.

# Why?

GoKit Cli by Kujtim Hoxha at https://github.com/kujtimiihoxha/kit is the only generator I can find besides the code generation efforts of go-kit itself at https://github.com/go-kit/kit/tree/master/cmd/kitgen.

GoKit Cli got me pretty far and I've learned a lot about go-kit using the Cli. I found though that the Cli would sometimes fail in ways I didn't understand or require some form of input I didn't understand. I chalk this up to not being familiar enough with Golang to understand any of the higher level jargon and concepts.

go-kit/cmd/kitgen is also far too advanced for me to use.

# @TODO:

- Support project generation
- Support dockerfile generation
- Support dev zipkin server working out of box
- Support api gateway generation
- Support middleware generation
- Support example project generation
- Support some degree of test generation
- Become useful
- Maybe rename from kitgen?

# Installation

This probably conflicts with go-kit/cmd/kitgen. I haven't tested.

This probably also requires a bunch of dependencies. I haven't tested.

This may not even work on non-Ubuntu platforms. I haven't tested.

```
go get github.com/allar/kitgen
cd $GOHOME/src/github.com/allar/kitgen
```

# Usage

## Initial Generation

Then inside a **empty** folder somewhere, run:

```
kitgen service -n foobar
go mod init
go run ./foobar/cmd
```

This will generate a service inside folder `foobar` named `foobar`, download its required packages, and start up the microservice.

By default, the service will be generated with two methods, `Foo` and `Bar`, both of which do nothing. To test if your server is running however, you can submit an empty post request to these methods by:

```
curl -d '' -X POST localhost:8081/foo
```

This should cause the service to log out an 'http server error', as it was unable to parse the posted data. If we give it a valid request object:

```
curl -d '{"s": "foo test"}' -X POST localhost:8081/foo
```

This should cause the service to log out that the Foo endpoint was called with response `{"rs":"","err":null}`

## Defining the Foo Service Method

Open `foobar/service/service.go` in your favorite text editor. You'll see that both Foo and Bar are defined to take in a string input named `s` and output a result string named `rs` as well as any error as `err`.

Open `foobar/service/service_gen.go` and you'll see that the default implementations of these methods is to, well, do nothing. Copy and paste the Foo method definition from `service_gen.go` into `service.go` to begin implementing it. For now, lets just take the input string and return the same string but in all caps as the result string. Your `service.go` should look something like:

```
package service

import (
    "context"
    "strings"
)

// Foobar describes the service interface
type FoobarService interface {
    // Add your methods here
        // e.x: Foo(ctx context.Context,s string)(rs string, err error)
    Foo(ctx context.Context,s string)(rs string, err error)
    Bar(ctx context.Context,s string)(rs string, err error)
}

// Foo implements business logic
// Cut and paste this function to your service.go to implement
func (is *implementedFoobarService) Foo(ctx context.Context, s string) (rs string, err error) {
        rs = strings.ToUpper(s)
        return rs, err
}
```

Once saved, it is important that you regenerate your service. From the parent folder of your foobar service (i.e. if your service is in `/my/stuff/foobar/` you should run the following within `/my/stuff/`:

```
kitgen generate -n foobar
```

This will regenerate your service, and remove the placeholder implementation of Foo inside `service_gen.go` if you accidentally left a copy in there or did not save it. Now that your service is regenerated, run it again with:

```
go run ./foobar/cmd
```

Now when you call a post to Foo, you should get back an uppercase string!

```
curl -d '{"s": "foo test"}' -X POST localhost:8081/foo
```

You should see that your service properly logs your method call using some logging middleware. Your service also exposes a prometheus metrics handler, which you can test by:

```
curl localhost:8080/metrics
```

The service also sets up zipkin tracing but I haven't got around to documenting it yet.

## Defining new service features

As long as your methods defined in your service interface are well formed, i.e. always take at least one parameter `ctx context.Context` as well as one result `err Error`, running service generation should create all the boilerplate you need. If you want to implement a stubbed method, move the method from `service_gen.go` to `service.go` and regenerate. Logging, instrumentation, and tracing middleware will be generated for all of your methods automatically.

# Code Generation

Assume all files ending in `_gen` will be wiped and regenerated when regenerating a service.
