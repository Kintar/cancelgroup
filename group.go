package cancelgroup

import (
	"context"
	"errors"
	"sync"
)

var ErrorParentContextCanceled = errors.New("group: parent context was canceled")

var errTaskAborted = errors.New("abort Task due to context completion")

// Task represents a function which will be run inside a group, but which is not capable of coordinating its work on
// a context.Context. A Task which is added to a group will be run inside a goroutine which will panic if the parent
// context is canceled. The panic will be recovered and will not terminate the runtime.
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
	// Go starts the given Task on a new goroutine.
	Go(Task)
}

// New creates a new Group and associated context.Context. If the calling code needs the ability to cancel a long-running
// Group, use NewWithContext and pass in a parent context for which the calling routine owns a context.CancelFunc or
// context.CancelCauseFunc.
//
// When all tasks run inside the returning group are complete, the Context will be completed and context.Cause will
// return nil if all tasks completed successful, or the error returned by the first failed task within the Group.
func New() (Group, context.Context) {
	ctx := context.Background()
	return NewWithContext(ctx), ctx
}

func NewWithContext(ctx context.Context) Group {
	g := &group{}
	g.ctx, g.cancelCause = context.WithCancelCause(ctx)
	return g
}

type group struct {
	wg          sync.WaitGroup
	errOnce     sync.Once
	ctx         context.Context
	cancelCause context.CancelCauseFunc
}

func (g *group) done() {
	g.wg.Done()
}

func (g *group) errored(err error) {
	g.errOnce.Do(func() { g.cancelCause(err) })
}

// recoverChildPanic traps a panic sequence from a Task in order to determine if the panic was caused by the Group
// context being canceled. If it was, we ignore the panic. If not, the panic is re-triggered.
func (g *group) recoverChildPanic() {
	if r := recover(); r != nil {
		//goland:noinspection GoTypeAssertionOnErrors
		if err, ok := r.(error); ok {
			if errors.Is(err, errTaskAborted) {
				g.errored(err)
				return
			}
		}
		panic(r)
	}
}

// Go adds the given Task to the Group and starts it on a new goroutine. If this Group's context is canceled, the
// goroutine will be terminated and the task will die in a nondeterministic state.
func (g *group) Go(t Task) {
	g.wg.Add(1)
	go func() {
		defer g.recoverChildPanic()
		if err := t(); err != nil {
			g.errored(err)
		}
	}()
}

// GoCtx adds the given CoordinatedTask to the Group and starts it on a new goroutine. If this Group's context is
// cancelled, the goroutine wil continue and the CoordinatedTask is expected to exit as soon as possible.
func (g *group) GoCtx(t CoordinatedTask) {
	panic("not implemented")
}

// Wait will block the calling goroutine until this Group's tasks have all completed successfully, or until the Group's
// context is canceled due to a child task error or the parent context's completion.
func (g *group) Wait() error {
	panic("not implemented")
}
