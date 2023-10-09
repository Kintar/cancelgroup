package cancelgroup

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	g := New()
	grp, ok := g.(*group)
	assert.Truef(t, ok, "New did not create *group")
	assert.NotNil(t, grp.cancelCause, "cancelCause")
	assert.NotNil(t, grp.ctxParent, "ctxParent")
	assert.NotNil(t, grp.waitFunction, "waitFunction")
}

func TestNewWithContext(t *testing.T) {
	pCtx, pCancel := context.WithCancelCause(context.Background())
	defer pCancel(nil)
	g := NewWithContext(pCtx)
	grp, ok := g.(*group)
	assert.Truef(t, ok, "NewWithContext did not create *group")
	assert.NotEqual(t, pCtx, grp.ctx, "contexts are equal")
}

func add(v *atomic.Int32) CoordinatedTask {
	return func(_ context.Context) error {
		v.Add(1)
		return nil
	}
}

func addNoCtx(v *atomic.Int32) Task {
	return func() error {
		v.Add(1)
		return nil
	}
}

const (
	TaskCount = 5
)

func Test_Go_and_Wait(t *testing.T) {
	var count atomic.Int32
	g := New()
	for i := 0; i < TaskCount; i++ {
		g.Go(add(&count))
	}

	assert.Nil(t, g.Wait())
	assert.EqualValues(t, TaskCount, count.Load())
}

func Test_Run_and_Wait(t *testing.T) {
	var count atomic.Int32
	g := New()
	for i := 0; i < TaskCount; i++ {
		g.Run(addNoCtx(&count))
	}

	assert.Nil(t, g.Wait())
	assert.EqualValues(t, TaskCount, count.Load())
}

func delayOnContext(ctx context.Context) error {
	<-time.After(time.Second)
	return nil
}

func delay() error {
	return delayOnContext(context.TODO())
}

func Test_waitReturnsAfterContextCancel(t *testing.T) {
	st := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	g := New()

	// this flag will be true once our waiting goroutine completes
	complete := false

	// this will delay for one second, but return immediately when the context is canceled
	g.Go(delayOnContext)

	// this will mark our flag when g.Wait() returns
	go func() {
		defer wg.Done()
		_ = g.Wait()
		complete = true
	}()

	// cancel the 1-second delay
	g.Cancel()
	// wait for the waiting goroutine to finish
	wg.Wait()

	// verify that complete was set
	assert.True(t, complete)
	// verify that it took less than one second, which means Wait returned before the dealy function fired
	assert.Greater(t, time.Second, time.Now().Sub(st))
}

func Test_waitReturnsAfterContextCancel_evenAfterRun(t *testing.T) {
	st := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	g := New()

	// this flag will be true once our waiting goroutine completes
	complete := false

	// this will delay for one second, but return immediately when the context is canceled
	g.Run(delay)

	// this will mark our flag when g.Wait() returns
	go func() {
		defer wg.Done()
		_ = g.Wait()
		complete = true
	}()

	// cancel the 1-second delay
	g.Cancel()
	// wait for the waiting goroutine to finish
	wg.Wait()

	// verify that complete was set
	assert.True(t, complete)
	// verify that it took less than one second, which means Wait returned before the delay function fired
	assert.Greater(t, time.Second, time.Now().Sub(st))
}

func Test_waitDoesNotReturnEarly(t *testing.T) {
	st := time.Now()
	g := New()

	// this will delay for one second, but return immediately when the context is canceled
	g.Go(func(_ context.Context) error {
		<-time.After(250 * time.Millisecond)
		return nil
	})

	assert.Nil(t, g.Wait())

	// verify that it took at least 250ms
	assert.LessOrEqual(t, 250*time.Millisecond, time.Now().Sub(st))
}

func Test_waitDoesNotReturnEarly_afterRun(t *testing.T) {
	st := time.Now()
	g := New()

	// this will delay for one second, but return immediately when the context is canceled
	g.Run(func() error {
		<-time.After(250 * time.Millisecond)
		return nil
	})

	assert.Nil(t, g.Wait())

	// verify that it took at least 250ms
	assert.LessOrEqual(t, 250*time.Millisecond, time.Now().Sub(st))
}

func Test_parentContextCancelsGroup(t *testing.T) {
	st := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cf := context.WithCancel(context.Background())
	g := NewWithContext(ctx)

	// this flag will be true once our waiting goroutine completes
	complete := false

	// this will delay for one second, but return immediately when the context is canceled
	g.Go(delayOnContext)

	// this will mark our flag when g.Wait() returns
	go func() {
		defer wg.Done()
		_ = g.Wait()
		complete = true
	}()

	// cancel the parent context
	cf()
	// wait for the waiting goroutine to finish
	wg.Wait()

	// verify that complete was set
	assert.True(t, complete)
	// verify that it took less than one second, which means Wait returned before the delay function fired
	assert.Greater(t, time.Second, time.Now().Sub(st))
}
