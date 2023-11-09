// Package pubsub provides a simple publisher/subscriber system that is type-safe and generic.
// Events are published and subscribed to according to type. Due to the nature of golangs type system,
// some types (slices, maps, and funcs) do not work with this system.
package pubsub

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

var (
	// Global is the default event scope. Publish and SubscribeTo use this event scope.
	Global *EventScope
)

func init() {
	Global = NewEventScope()
}

// EventScope describes a scope where publishers and subscribers communicate with each other.
// Data published to an event scope will not be seen by subscribers listening on a differnt
// event scope. Multiple event scopes should only be used when you need to publish data with
// the same type but different handlers.
type EventScope struct {
	subscribers *sync.Map
}

// UnSubFn is a function which unsubscribes from the data type. Calling this will close the
// channel returned by SubscribeTo/SubscribeToScope.
type UnsubFn func()

func NewEventScope() *EventScope {
	return &EventScope{
		subscribers: &sync.Map{},
	}
}

// Publish will send the value val into the global event scope. If the context is canceled,
// the value may not be sent to all subscribers.
func Publish[T any](ctx context.Context, val T) {
	PublishToScope(ctx, Global, val)
}

// PublishToScope will send the value val on the specified event scope. If the context is canceled,
// the value may not be sent to all subscribers.
func PublishToScope[T any](ctx context.Context, e *EventScope, val T) {
	var zero T
	subs, ok := e.subscribers.Load(zero)
	if !ok {
		return
	}

	subMap := subs.(*sync.Map)
	subMap.Range(func(_, value any) bool {
		go func() {
			dest := value.(chan any)
			select {
			case dest <- val:
			case <-ctx.Done():
				return
			}

		}()
		return true
	})
}

// SubscribeTo creates a channel to listen for events of type T. When listeners are finished
// processing these events, the UnsubFn should be called.
func SubscribeTo[T any](ctx context.Context) (chan T, UnsubFn) {
	return SubscribeToScope[T](ctx, Global)
}

// SubscribeTo creates a channel to listen for events of type T published on the provided event scope.
// When listeners are finished processing these events, the UnsubFn should be called.
func SubscribeToScope[T any](ctx context.Context, e *EventScope) (chan T, UnsubFn) {
	ch := make(chan T)
	untypedCh := make(chan any)
	id := uuid.New()

	var zero T

	// This line can panic if a non-hashable value is passed in
	subs, _ := e.subscribers.LoadOrStore(zero, &sync.Map{})
	subMap := subs.(*sync.Map)

	subMap.Store(id, untypedCh)

	forwardCtx, cancel := context.WithCancel(ctx)
	go castAndForward(forwardCtx, untypedCh, ch)

	unsub := func() {
		subMap.Delete(id)
		cancel()
	}

	return ch, unsub
}

func castAndForward[T any](ctx context.Context, in <-chan any, out chan<- T) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-in:
			if !ok {
				return
			}
			typedVal, ok := val.(T)
			if !ok {
				panic("mismatched type")
			}
			select {
			case out <- typedVal:
			case <-ctx.Done():
				return
			}
		}
	}
}
