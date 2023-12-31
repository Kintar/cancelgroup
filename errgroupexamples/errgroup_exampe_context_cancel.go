package errgroupexamples

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

// ParentContextCancel demonstrates that canceling the parent context of an errgroup.Group does not immediately abort
// running goroutines within that group.
func ParentContextCancel() {
	var gr1Complete, gr2Complete bool
	gr1 := func() error {
		<-time.After(time.Second * 4)
		gr1Complete = true
		return nil
	}
	gr2 := func() error {
		<-time.After(time.Second * 2)
		gr2Complete = true
		return nil
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	g, _ := errgroup.WithContext(ctx)
	g.Go(gr1)
	g.Go(gr2)
	cancel(errors.New("parent canceled"))
	<-ctx.Done()
	fmt.Println("After context cancel:", gr1Complete, gr2Complete, context.Cause(ctx))
	err := g.Wait()
	fmt.Println("After group.Wait: ", gr1Complete, gr2Complete, err)
}

// ErrGroupCancel demonstrates that g.Wait() does not return until all goroutines have exited, regardless of the group's
// context's cancellation status.
func ErrGroupCancel() {
	gr1 := func() error {
		<-time.After(time.Second * 4)
		return errors.New("group 1 error")
	}
	gr2 := func() error {
		<-time.After(time.Second * 2)
		return errors.New("group 2 error")
	}
	start := time.Now()
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(gr1)
	g.Go(gr2)
	<-ctx.Done()
	fmt.Printf("context has been canceled after %dms: %v\n", time.Now().Sub(start).Milliseconds(), context.Cause(ctx))
	err := g.Wait()
	fmt.Printf("group.Wait is now complete after %dms: %v\n", time.Now().Sub(start).Milliseconds(), err)
}
