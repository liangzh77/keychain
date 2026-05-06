package server

import (
	"net/http"
	"time"
)

type Options struct {
	Now func() time.Time
}

func NewRouter(options Options) http.Handler {
	if options.Now == nil {
		options.Now = time.Now
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(options.Now))
	return withJSONHeaders(mux)
}

func withJSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
