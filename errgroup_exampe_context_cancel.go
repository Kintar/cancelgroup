package cancelgroup

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

// Cancel demonstrates that canceling the parent context of an errgroup.Group does not immediately
// abort running goroutines within that group.
//

func ExampleGroup_cancel() {
	var gr1_complete, gr2_complete bool
	gr1 := func() error {
		<-time.After(time.Second * 4)
		gr1_complete = true
		return nil
	}
	gr2 := func() error {
		<-time.After(time.Second * 2)
		gr2_complete = true
		return nil
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	g, _ := errgroup.WithContext(ctx)
	g.Go(gr1)
	g.Go(gr2)
	cancel(errors.New("parent canceled"))
	<-ctx.Done()
	fmt.Println("After context cancel:", gr1_complete, gr2_complete, context.Cause(ctx))
	err := g.Wait()
	fmt.Println("After Group.Wait: ", gr1_complete, gr2_complete, err)
}
