package cancelgroup_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kintar/cancelgroup"
	"math/rand"
	"sync"
	"time"
)

var ErrorWorkerCanceled = errors.New("worker: cancel")

// DelayWorker creates a CoordinatedTask which simply waits for a random delay
func DelayWorker() cancelgroup.CoordinatedTask {
	return func(ctx context.Context) error {
		delay := time.After(time.Duration(rand.Intn(250)) * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				return ErrorWorkerCanceled
			case <-delay:
				return nil
			}
		}
	}
}

// Co schedules a CoordinatedTask on the Group. Coordinated tasks are simply functions which accept a context.Context
// as their only parameter, and are expected to exit gracefully as soon as possible after the context is canceled.
func ExampleGroup_Co() {
	g := cancelgroup.New()

	g.Co(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	g.Go(func() error {
		<-time.After(time.Millisecond * 250)
		g.Cancel()
		return nil
	})

	fmt.Println(g.Wait())
}

// Go schedules a Task on the Group. Tasks are simple functions which return an error, but are not capable of exiting
// based on the Done status of a context.Context. They will continue to run even after a Group is canceled, but will
// not affect the value returned by Group.Wait().
func ExampleGroup_Go() {
	g := cancelgroup.New()
	errTaskCompleted := errors.New("task completed")

	// a sync.WaitGroup so we can coordinate the final message of this example with the completion of our scheduled Task
	var wg sync.WaitGroup
	wg.Add(1)

	// track the start time of our Group
	start := time.Now()

	g.Go(func() error {
		defer wg.Done()
		<-time.After(time.Millisecond * 250)
		return errTaskCompleted
	})

	// cancel the Group so that Wait returns immediately
	g.Cancel()

	// read the error
	err := g.Wait()

	// track the time the Group.Wait call completed
	waitCompleteAt := time.Now()

	fmt.Println("wait completed in", waitCompleteAt.Sub(start).Milliseconds(), "ms")

	// err will be ErrorGroupCanceled, not errTaskCompleted
	fmt.Println("Group error is: ", err)

	// Now wait for the scheduled task to finish
	wg.Wait()
	taskCompleteAt := time.Now()

	fmt.Println("task completed in", taskCompleteAt.Sub(start).Milliseconds(), "ms")

	// err will be ErrorGroupCanceled, not errTaskCompleted
	fmt.Println("Group error is: ", err)
}
