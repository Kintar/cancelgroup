/*
Package cancelgroup expands on the capabilities of golang.org/x/sync/errgroup to provide fail-fast cancellation of routines running within a group.

See the [cancellation examples] for an explanation of the context cancel behavior of errgroup.Group. The purpose of the
cancelgroup package is to allow g.Wait() to return immediately upon cancellation of the parent context, and when possible
to abort the execution of the subordinate goroutines.

[cancellation examples]: https://github.com/Kintar/cancelgroup/
*/
package cancelgroup
