# cancelgroup

[![PkgGoDev](https://pkg.go.dev/badge/github.com/Kintar/cancelgroup)](https://pkg.go.dev/github.com/Kintar/cancelgroup)

This package provides `cancelgroup.Group`, which provides an interface similar to [errgroup.Group](https://pkg.go.dev/golang.org/x/sync/errgroup)
with the additional ability to cancel the Group's context and cause `Wait` to return immediately.

Further explanation is available in the package documentation, linked above.