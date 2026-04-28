// server/router.go
//
// Layer 1: Router
//
// The router maps (method, path) pairs to handler functions.
// It's a simplified version of http.ServeMux — but explicit about
// what matching logic it's doing.
//
// After Layer 2, we'll swap this out for net/http's ServeMux and
// reflect on what we had to write ourselves vs what we got for free.

package server

import (
	"strings"
)

// RouteHandlerFunc handles a request and returns a response.
// It receives a params map for URL path parameters like /:code.
type RouteHandlerFunc func(req Request, params map[string]string) Response

// route is a single registered route entry.
type route struct {
	method  string           // "GET", "POST", etc.
	pattern string           // "/shorten", "/:code", "/stats/:code"
	handler RouteHandlerFunc
}

// Router dispatches incoming requests to the appropriate handler.
//
// How matching works:
//  1. Iterate routes in registration order
//  2. Check if the pattern matches the request path (including path params)
//  3. If path matches but method doesn't → track it (for 405 response)
//  4. If both match → call the handler
//  5. If nothing matches → 404
//
// This is O(n) per request — fine for a small number of routes.
// net/http's ServeMux uses a map + longest-prefix matching for O(1).
type Router struct {
	routes []route
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// GET registers a handler for GET requests on pattern.
func (r *Router) GET(pattern string, handler RouteHandlerFunc) {
	r.routes = append(r.routes, route{method: "GET", pattern: pattern, handler: handler})
}

// POST registers a handler for POST requests on pattern.
func (r *Router) POST(pattern string, handler RouteHandlerFunc) {
	r.routes = append(r.routes, route{method: "POST", pattern: pattern, handler: handler})
}

// Handler returns a HandlerFunc that dispatches through this router.
// This is the bridge between the Server (which takes a HandlerFunc) and the
// Router (which has multiple registered routes).
//
// Usage:
//
//	router := server.NewRouter()
//	router.GET("/:code", handlers.Redirect)
//	srv := server.New(":8080", router.Handler())
func (r *Router) Handler() HandlerFunc {
	return func(req Request) Response {
		return r.dispatch(req)
	}
}

// dispatch finds the matching route for a request and calls its handler.
func (r *Router) dispatch(req Request) Response {
	// Strip query string before matching: "/foo?bar=baz" → "/foo"
	path := req.Path
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}

	var pathMatched bool // did any route match the path (regardless of method)?

	for _, rt := range r.routes {
		params, ok := matchPath(rt.pattern, path)
		if !ok {
			continue // pattern doesn't match this path at all
		}
		pathMatched = true

		if rt.method != req.Method {
			continue // path matched but wrong method — keep looking
		}

		return rt.handler(req, params)
	}

	// If a path matched but no method did, return 405 not 404.
	// This matters: 404 means "this route doesn't exist",
	// 405 means "this route exists but not for your method".
	if pathMatched {
		return Response{
			StatusCode: 405,
			StatusText: "Method Not Allowed",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"error":"method not allowed"}`),
		}
	}

	return Response{
		StatusCode: 404,
		StatusText: "Not Found",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"error":"not found"}`),
	}
}

// matchPath checks whether a request path matches a route pattern.
//
// Pattern syntax:
//   - Literal segments must match exactly:   "/shorten" matches "/shorten"
//   - Named params start with ":":           "/:code" matches "/abc123"
//                                            → params["code"] = "abc123"
//
// Examples:
//
//	matchPath("/shorten", "/shorten")       → {}, true
//	matchPath("/:code", "/abc123")          → {"code": "abc123"}, true
//	matchPath("/stats/:code", "/stats/x1")  → {"code": "x1"}, true
//	matchPath("/:code", "/a/b")             → nil, false  (segment count mismatch)
//	matchPath("/shorten", "/other")         → nil, false
func matchPath(pattern, path string) (map[string]string, bool) {
	patternSegs := splitPath(pattern)
	pathSegs := splitPath(path)

	// Segment count must match — we don't support wildcards
	if len(patternSegs) != len(pathSegs) {
		return nil, false
	}

	params := make(map[string]string)

	for i, seg := range patternSegs {
		if strings.HasPrefix(seg, ":") {
			// Named parameter — capture the corresponding path segment
			paramName := seg[1:] // strip leading ":"
			params[paramName] = pathSegs[i]
		} else if seg != pathSegs[i] {
			// Literal segment mismatch
			return nil, false
		}
	}

	return params, true
}

// splitPath splits a URL path into non-empty segments.
//
//	"/"             → []
//	"/shorten"      → ["shorten"]
//	"/stats/abc123" → ["stats", "abc123"]
//	"/a//b/"        → ["a", "b"]   (double slashes and trailing slash handled)
func splitPath(path string) []string {
	raw := strings.Split(path, "/")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}