package iter

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Mapper can be used to configure the behaviour of iterators that
// have a result type R, such as Map and MapErr. The zero value is
// safe to use with reasonable defaults.
//
// Mapper is also safe for reuse and concurrent use.
type Mapper[T, R any] struct {
	// Iterator is the underlying implementation of Mapper methods.
	// It can be used to configure the maximum number of goroutines
	// that Mapper methods can use.
	Iterator[T]
}

// Map applies f to each element of input, returning the mapped result.
//
// Map uses Iterator to perform the iteration, which always uses at most
// runtime.GOMAXPROCS goroutines. For a configurable goroutine limit,
// use a custom Mapper.
func Map[T, R any](input []T, f func(*T) R) []R {
	return Mapper[T, R]{}.Map(input, f)
}

// Map applies f to each element of input, returning the mapped result.
//
// Map uses Iterator to perform the iteration, using up to the configured
// Iterator's maximum number of goroutines.
func (m Mapper[T, R]) Map(input []T, f func(*T) R) []R {
	res := make([]R, len(input))
	m.Iterator.ForEachIdx(input, func(i int, t *T) {
		res[i] = f(t)
	})
	return res
}

// MapErr applies f to each element of the input, returning the mapped result
// and a combined error of all returned errors.
//
// MapErr uses Iterator to perform the iteration, which always uses at
// most runtime.GOMAXPROCS goroutines. For a configurable goroutine limit,
// use a custom Mapper.
func MapErr[T, R any](input []T, f func(*T) (R, error)) ([]R, error) {
	return Mapper[T, R]{}.MapErr(input, f)
}

// MapErr applies f to each element of the input, returning the mapped result
// and a combined error of all returned errors.
//
// MapErr uses Iterator to perform the iteration, using up to the configured
// Iterator's maximum number of goroutines.
func (m Mapper[T, R]) MapErr(input []T, f func(*T) (R, error)) ([]R, error) {
	var (
		res    = make([]R, len(input))
		errMux sync.Mutex
		errs   error
	)
	m.Iterator.ForEachIdx(input, func(i int, t *T) {
		var err error
		res[i], err = f(t)
		if err != nil {
			errMux.Lock()
			// TODO: use stdlib errors once multierrors land in go 1.20
			errs = errors.Append(errs, err)
			errMux.Unlock()
		}
	})
	return res, errs
}
