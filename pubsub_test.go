package pubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubSub(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[int](ctx, testScope)
	defer unsub()

	val := 42
	PublishToScope(ctx, testScope, val)

	incVal, ok := <-testingCh

	assert.True(t, ok)
	assert.Equal(t, val, incVal)
}

func TestPubSub_Ptr(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[*string](ctx, testScope)
	defer unsub()

	str := "my string"
	val := &str
	PublishToScope(ctx, testScope, val)

	incVal, ok := <-testingCh

	assert.True(t, ok)
	assert.Equal(t, val, incVal)
}

func TestPubsub_Struct(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	type testStruct struct {
		foo int
		bar string
	}

	testingCh, unsub := SubscribeToScope[testStruct](ctx, testScope)
	defer unsub()

	val := testStruct{foo: 42, bar: "test"}

	PublishToScope(ctx, testScope, val)

	incVal, ok := <-testingCh

	assert.True(t, ok)
	assert.Equal(t, val, incVal)
}

func TestPubSub_SlicePanics(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	success := false
	defer func() {
		recover()
		if success {
			t.FailNow()
		}
	}()

	SubscribeToScope[[]bool](ctx, testScope)

	success = true
}

func TestPubSub_MapPanics(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	success := false
	defer func() {
		recover()
		if success {
			t.FailNow()
		}
	}()

	SubscribeToScope[map[any]any](ctx, testScope)

	success = true
}

func TestPubSub_Chan(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[chan any](ctx, testScope)
	defer unsub()

	val := make(chan any)

	PublishToScope(ctx, testScope, val)

	incVal, ok := <-testingCh

	assert.True(t, ok)
	assert.Equal(t, val, incVal)
}

func TestPubSub_FnPanics(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	success := false
	defer func() {
		recover()
		if success {
			t.FailNow()
		}
	}()

	SubscribeToScope[func()](ctx, testScope)

	success = true
}

type testInterface interface {
	Test()
}

type testImpl struct {
}

func (t testImpl) Test() {}

func TestPubSub_Intf(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[testInterface](ctx, testScope)
	defer unsub()

	val := testImpl{}
	PublishToScope[testInterface](ctx, testScope, val)

	incVal, ok := <-testingCh
	assert.True(t, ok)
	assert.Equal(t, val, incVal)
}

func TestPubSub_Unsub(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[int](ctx, testScope)

	unsub()

	_, ok := <-testingCh
	assert.False(t, ok)
}

// This test only fails if it panics
func TestPubSub_NoSub(t *testing.T) {
	ctx := context.Background()
	testScope := NewEventScope()

	PublishToScope(ctx, testScope, 1)
}

func TestPubSub_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	testScope := NewEventScope()

	testingCh, unsub := SubscribeToScope[int](ctx, testScope)
	defer unsub()
	cancel()

	_, ok := <-testingCh
	assert.False(t, ok)
}
