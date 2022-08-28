// Package web contains a small web framework extension.
package web

import (
	"context"
	"fmt"
	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"syscall"
	"time"
)

// Handler is a type that handles a http request within our own little mini
// framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	mux      *httptreemux.ContextMux
	otmux    http.Handler
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp create an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux := httptreemux.NewContextMux()

	return &App{
		mux:      mux,
		otmux:    otelhttp.NewHandler(mux, "request"),
		shutdown: shutdown,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up on line 44 in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

// Handle sets a Handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		//Pull the context from the request and
		// use it as a separate parameter.
		ctx := r.Context()

		// Capture the parent request span from the context.
		span := trace.SpanFromContext(ctx)

		// Set the context with the required values to
		// process the request
		v := Values{
			Now:     time.Now().UTC(),
			TraceID: span.SpanContext().TraceID().String(),
		}
		ctx = context.WithValue(ctx, key, &v)

		if err := handler(ctx, w, r); err != nil {
			// ERROR HANDLING

			a.SignalShutdown()
			return
		}

		// POST CODE PROCESSING
	}

	finalPath := path
	if group != "" {
		finalPath = fmt.Sprintf("/%s%s", group, path)
	}
	a.mux.Handle(method, finalPath, h)
}
