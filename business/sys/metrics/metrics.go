package metrics

import (
	"context"
	"expvar"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package already based on a singleton
// for the different metrics that are registered with the package so there
// isn't much choice here.
var m *metrics

// metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently thanks to expvar. No extra abstraction is required.
type metrics struct {
	goroutine *expvar.Int
	requests  *expvar.Int
	errors    *expvar.Int
	panics    *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable since everything
// inside of expvar is registered as a singleton. The use of once will make
// sure this initialization only happens once.
func init() {
	m = &metrics{
		goroutine: expvar.NewInt("goroutines"),
		requests:  expvar.NewInt("requests"),
		errors:    expvar.NewInt("errors"),
		panics:    expvar.NewInt("panics"),
	}
}

type ctxKey int

const key ctxKey = 1

// Set sets the metrics data into the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// AddGoroutines increments the goroutine metric by 1.
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		if v.goroutine.Value()%100 == 0 {
			v.goroutine.Add(1)
		}
	}
}

// AddRequests increments the request metric by 1.
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.requests.Add(1)
	}
}

// AddErrors increments the error metric by 1.
func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.errors.Add(1)
	}
}

// AddPanics increments the panics metric by 1.
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.panics.Add(1)
	}
}
