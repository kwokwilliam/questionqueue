package handlers

import "net/http"

const accessControlAllowOrigin = "Access-Control-Allow-Origin"
const accessControlAllowMethods = "Access-Control-Allow-Methods"
const accessControlAllowHeaders = "Access-Control-Allow-Headers"
const accessControlExposedHeaders = "Access-Control-Expose-Headers"
const accessControlMaxAge = "Access-Control-Max-Age"

const allowedMethods = "GET, PUT, POST, PATCH, DELETE"
const allowedHeaders = "Content-Type, Authorization"
const exposedHeaders = "Authorization"
const maxAge = "600"

// CORS is a middleware handler that sets CORS headers
type CORS struct {
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the real handler
// after adding CORS headers to everything
func (c *CORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(accessControlAllowOrigin, allOrigins)
	w.Header().Set(accessControlAllowMethods, allowedMethods)
	w.Header().Set(accessControlAllowHeaders, allowedHeaders)
	w.Header().Set(accessControlExposedHeaders, exposedHeaders)
	w.Header().Set(accessControlMaxAge, maxAge)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	c.handler.ServeHTTP(w, r)
}

// NewCORS constructs a new CORS middleware handler
func NewCORS(handlerToWrap http.Handler) *CORS {
	return &CORS{handlerToWrap}
}
