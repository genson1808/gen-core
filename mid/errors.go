package mid

import (
	"context"
	"gitlab.com/genson1808/gen-core/validate"
	"gitlab.com/genson1808/gen-core/web"
	"go.uber.org/zap"
	"net/http"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {

	// This is handler actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated error.
			if err := handler(ctx, w, r); err != nil {

				// Log the error.
				log.Errorw("ERROR", "traceid", v.TraceID, "message", err)

				// Build out the error response.
				var er web.ErrorResponse
				var status int
				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = web.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case web.IsRequestError(err):
					reqErr := web.GetRequestError(err)
					er = web.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = web.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Response with the error back to the client.
				if err := web.Response(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down service.
				if web.IsShutdown(err) {
					return err
				}
			}

			// The error hash been handled so we can stop propagating it.
			return nil
		}

		return h
	}
	return m
}
