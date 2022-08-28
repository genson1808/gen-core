package metrics

import (
	"context"
	"expvar"
	"runtime"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package is already based on a singleton
// for the different metrics that are registered with the package so there
// isn't much choice here.
var m *metrics

// metrics represents the set of metrics
type metrics struct {
	goroutines *expvar.Int
	request    *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable since everything
// inside of expvar is registered as singleton. The use of once will make
// sure this initialization only happens once.
func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		request:    expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

// Metrics will be supported through the context.

// ctxKeyMetrics represents the type of value for the context key.
type ctxKey int

// key is how metrics values are stored/retrieved.
const key ctxKey = 1

// Set sets the metrics data into the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// Add more of these functions when a metric needs to be collected in
// different parts of the codebase. This will keep this package the
// central authority for metrics and metrics won't get lost.

// AddGoroutines refreshes the goroutine metric every 100 request.
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		if v.request.Value()%100 == 0 {
			v.goroutines.Set(int64(runtime.NumGoroutine()))
		}
	}
}

// AddRequests increments the request metrics by 1.
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.request.Add(1)
	}
}

// AddErrors increments the errors metrics by 1.
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
