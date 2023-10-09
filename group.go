package cancelgroup

import (
	"context"
	"errors"
	"sync"
)

// ErrorGroupCanceled is returned from the Wait or ContextCause methods of a Group when the Group's Cancel method is
// called.
var ErrorGroupCanceled = errors.New("Group canceled")

// ErrorParentContextCanceled is returned from the Wait or ContextCause methods of a Group which was aborted due to its
// parent context being canceled.
var ErrorParentContextCanceled = errors.New("Group parent context was canceled")

// Task represents a function which will be run inside a Group, but which is not capable of coordinating its work on
// a context.Context. A Task which is run in a Group will continue to run until it naturally terminates, even if the
// Group or its parent context is canceled.
type Task func() error

// CoordinatedTask represents a Task which is capable of monitoring a context and cleanly exiting if the context is
// canceled prior to the Task's execution.
type CoordinatedTask func(ctx context.Context) error

// New creates a new Group. If the calling code needs the ability to cancel a long-running Group, use NewWithContext and
// pass in a parent context for which the calling routine owns a context.CancelFunc or context.CancelCauseFunc.
func New() *Group {
	return NewWithContext(context.Background())
}

// NewWithContext creates a new Group and derives a new context from the given Context. If the parent context is canceled,
// the Group's context will also be canceled.
//
// When all tasks run inside the returning Group are complete, the Context will be completed. The Group context's casuse
// will be nil if all tasks completed successfully, otherwise it will contain the first error returned by a Group task.
func NewWithContext(ctx context.Context) *Group {
	g := &Group{
		ctxParent: ctx,
	}

	g.waitFunction = sync.OnceValue(func() error { return wait(g) })

	g.ctx, g.cancelCause = context.WithCancelCause(ctx)
	return g
}

// Group provides synchronization and cancellation for related goroutines working on a shared task.
type Group struct {
	wg           sync.WaitGroup
	errOnce      sync.Once
	waitFunction func() error
	ctxParent    context.Context
	ctx          context.Context
	cancelCause  context.CancelCauseFunc
}

func (g *Group) done() {
	g.wg.Done()
}

func (g *Group) errored(err error) {
	g.errOnce.Do(func() { g.cancelCause(err) })
}

func (g *Group) Cancel() {
	g.errored(ErrorGroupCanceled)
}

// Go adds the given Task to the Group and starts it on a new goroutine. Since a Task does not accept a context, Tasks
// will continue to run until their natural termination even if the Group is canceled.
func (g *Group) Go(t Task) {
	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := t(); err != nil {
			g.errored(err)
		}
	}()
}

// Co adds the given CoordinatedTask to the Group and starts it on a new goroutine, passing in the Group's context.
// A CoordinatedTask **MUST** monitor ctx.Done() and exit as soon as possible once detecting the context is canceled.
func (g *Group) Co(t CoordinatedTask) {
	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := t(g.ctx); err != nil {
			g.errored(err)
		}
	}()
}

type nothing struct{}

// Wait will block the calling goroutine until this Group's tasks have all completed successfully, or until the Group's
// context is canceled due to a child task error or the parent context's completion, whichever happens first. Once Wait
// returns, further calls will be non-blocking and will return the same value.
func (g *Group) Wait() error {
	return g.waitFunction()
}

func wait(g *Group) error {
	ch := make(chan nothing)
	go func() {
		g.wg.Wait()
		ch <- nothing{}
	}()
	select {
	case <-g.ctx.Done():
		g.errored(ErrorParentContextCanceled)
	case <-ch:
	}

	return context.Cause(g.ctx)
}
