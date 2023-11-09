# Pubsub
[![GoDoc](https://pkg.go.dev/badge/github.com/WillYingling/pubsub?status.svg)](https://pkg.go.dev/github.com/WillYingling/pubsub?tab=doc)

This package is a simple publisher-subscriber library in Golang. It takes advantage of the generics added in Go1.18
to create a pubsub system keyed by types. This removes some of the type assertions that typically come with systems that rely on `any`. 

## Non-generic pubsub systems

Let's say we are trying to pass around a struct `UserEvent`.
A typical pubsub system will look something like this:

Publishing code:
```go
// Payload is usually type any
event := pubsub.Event{
    Payload: UserEvent{}
}

// UserEventTopic is some arbitrary constant to control what subscribers
// receive this event.
pubsub.Publish(UserEventTopic, event)
```

Subscribing code:
```go
// Setting up the subscriber
// Use the same constant for the subscription code
eventCh, unsub := pubsub.Subscribe(UserEventTopic)
defer unsub()

// Handling events
event := <- eventCh

// Type cast back to the type we need
userEvent, ok := event.Payload.(UserEvent)

if !ok {
    // Handle incompatible type
}
```

This system lacks type safety, there's nothing that prevents calling code
from passing other data types on the topic `UserEventTopic`. The system
needs to maintain a list of arbitrary topics and their associated types.

## Examples

With this library, the topics are the types themselves, so no external topic
list needs to be maintained and the type assertions are built into the 
pubsub system. Lets see the same example using this package:

Publishing code:

```go
userEvent := UserEvent{}

// No need to wrap the data in a new type or specify an arbitrary constant.
pubsub.Publish(ctx, userEvent)
```

Subscribing code:
```go
// Setting up the subscriber
userEventCh, unsub := pubsub.SubscribeTo[UserEvent](ctx)
defer unsub()

// Handling events
// Values come out of the channel with the correct type
userEvent := <- userEventCh
```

### Working with interface types

Publishing to interface types is more complicated since they have an
underlying type.

```go
// Subscribe to an interface type
errCh, unsub := pubsub.SubscribeTo[error](ctx)
```

```go
// Some struct that implements the error interface
err := userError{}

// BAD, this will publish to the userError topic
pubsub.Publish(ctx, err)

// GOOD, this publishes to the error topic
pubsub.Publish[error](ctx, err)
```

## Limitations

Keying topics by type has some limitations introduced by the golang type
system. Specifically, non-hashable types cannot be subscribed to. Here is
a comprehensive list of golang types and whether or not they work with this pubsub system.

As of golang 1.21.3:
| Type | Compatibility | Workaround |
|:-|:-:|:-|
| Primitives | Yes | |
| Pointers | Yes | |
| Structs | Yes | |
| Slices | No | Wrap the slice in a struct type |
| Maps | No | Wrap the map in a struct type |
| Channels | Yes | |
| Functions | No | Use an interface type instead |
| Interfaces | Yes* | See interface example |