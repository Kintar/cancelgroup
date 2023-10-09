/*
Package cancelgroup provides a mechanism to coordinate related tasks of goroutines and associate them with a cancelable
context.

The [errgroup package] is very useful for coordinating groups of goroutines and limiting the number of active goroutines,
and will allow other goroutines to coordinate based on a context created by the error Group. However, there is no way to
supply a parent context which can cancel the Group, and the Group's `Wait()` method blocks until all goroutines are complete
even when a routine fails and sets the Group's error.

The Group in this package guarantees that Wait will return immediately when the Group's context is canceled. It also
provides a mechanism to schedule tasks which accept an incoming context, allowing those tasks to cleanly exit if the
Group is canceled.

[errgroup package]: https://pkg.go.dev/golang.org/x/sync/errgroup
*/
package cancelgroup
