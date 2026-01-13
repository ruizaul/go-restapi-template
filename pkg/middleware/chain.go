// Package middleware provides HTTP middleware functions for the API.
package middleware

import "net/http"

// Chain wraps a handler with multiple middlewares.
// Middlewares are applied in the order they are provided (first middleware is outermost).
//
// Example:
//
//	handler := middleware.Chain(
//	    myHandler,
//	    middleware.Logging(logger),
//	    middleware.Recovery(logger),
//	    middleware.CORS(corsConfig),
//	)
//
// This is equivalent to: Logging(Recovery(CORS(myHandler)))
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply middlewares in reverse order so the first middleware in the list
	// is the outermost (first to receive the request, last to send the response)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ChainFunc is like Chain but accepts an http.HandlerFunc instead of http.Handler.
func ChainFunc(handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return Chain(handler, middlewares...)
}
