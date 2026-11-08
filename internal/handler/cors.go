package handler

import "net/http"

// WithCORS answers preflights and sets permissive CORS headers.
//
// The API authenticates with a Bearer token in the Authorization header, never
// with cookies, and the web app calls the API from its own server, so browsers
// never need to send credentialed cross-origin requests to it. For that reason
// Authorization is intentionally NOT listed in Access-Control-Allow-Headers: a
// browser page on another origin cannot send the token, and the wildcard
// origin is therefore harmless.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
