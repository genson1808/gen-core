package web

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

func Response(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "foundation.web.respond")
	span.SetAttributes(attribute.Int("statusCode", statusCode))
	defer span.End()

	// Set the status code for the request logger middleware.
	SetStatusCode(ctx, statusCode)

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//Set the content once we know marshaling has succeeded.
	w.Header().Set("content-Type", "application/json")

	// Write the result status code to the response.
	w.WriteHeader(statusCode)

	// Sen the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
