package cancelgroup

import (
	"context"
	"errors"
	"sync"
)

// ErrorGroupCanceled is returned from the Wait or ContextCause methods of a Group when the group's Cancel method is
// called.
var ErrorGroupCanceled = errors.New("group canceled")

// ErrorParentContextCanceled is returned from the Wait or ContextCause methods of a Group which was aborted due to its
// parent context being canceled.
var ErrorParentContextCanceled = errors.New("group parent context was canceled")

// Task represents a function which will be run inside a group, but which is not capable of coordinating its work on
// a context.Context. A Task which is run in a group will continue to run until it naturally terminates, even if the
// Group or its parent context is canceled.
type Task func() error

// CoordinatedTask represents a Task which is capable of monitoring a context and cleanly exiting if the context is
// canceled prior to the Task's execution.
type CoordinatedTask func(ctx context.Context) error

// Group represents a coordinated collection of goroutines working on subtasks of a common task.
//
// If created with the NewWithContext function, the resulting Group will be aborted when the parent context is
// canceled, and the Group's Wait() method will return ErrorParentContextCanceled.
type Group interface {
	// Wait blocks the calling goroutine until this Group completes, then returns the error, if any, which aborted
	// the group.
	Wait() error
	// Run starts the given Task on a new goroutine. Since the Task deos not accept a context, it is not capable of
	// being canceled.
	Run(Task)
	// Go starts the given CoordinatedTask on a new goroutine, passing the context of the Group in for coordination of
	// early cancel.
	Go(task CoordinatedTask)
	// Cancel cancels the group's context and sets
	Cancel()
}

// New creates a new Group. If the calling code needs the ability to cancel a long-running Group, use NewWithContext and
// pass in a parent context for which the calling routine owns a context.CancelFunc or context.CancelCauseFunc.
func New() Group {
	return NewWithContext(context.Background())
}

// NewWithContext creates a new Group and derives a new context from the given Context. If the parent context is canceled,
// the Group's context will also be canceled.
//
// When all tasks run inside the returning group are complete, the Context will be completed. The group context's casuse
// will be nil if all tasks completed successfully, otherwise it will contain the first error returned by a group task.
func NewWithContext(ctx context.Context) Group {
	g := &group{
		ctxParent: ctx,
	}

	g.waitFunction = sync.OnceValue(func() error { return wait(g) })

	g.ctx, g.cancelCause = context.WithCancelCause(ctx)
	return g
}

type group struct {
	wg           sync.WaitGroup
	errOnce      sync.Once
	waitFunction func() error
	ctxParent    context.Context
	ctx          context.Context
	cancelCause  context.CancelCauseFunc
}

func (g *group) done() {
	g.wg.Done()
}

func (g *group) errored(err error) {
	g.errOnce.Do(func() { g.cancelCause(err) })
}

func (g *group) Cancel() {
	g.errored(ErrorGroupCanceled)
}

// Run adds the given Task to the Group and starts it on a new goroutine. Since a Task does not accept a context, Tasks
// will continue to run until their natural termination even if the Group is canceled.
func (g *group) Run(t Task) {
	g.wg.Add(1)
	go func() {
		defer g.done()
		if err := t(); err != nil {
			g.errored(err)
		}
	}()
}

// Go adds the given CoordinatedTask to the Group and starts it on a new goroutine, passing in the Group's context.
// A CoordinatedTask **MUST** monitor ctx.Done() and exit as soon as possible once detecting the context is canceled.
func (g *group) Go(t CoordinatedTask) {
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
// context is canceled due to a child task error or the parent context's completion.
func (g *group) Wait() error {
	return g.waitFunction()
}

func wait(g *group) error {
	ch := make(chan nothing)
	go func() {
		g.wg.Wait()
		ch <- nothing{}
	}()
	select {
	case <-g.ctx.Done():
	case <-ch:
	}

	return context.Cause(g.ctx)
}
