package server

import (
	"context"
	"net/http"
	"time"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/auth"
)

type Options struct {
	Now         func() time.Time
	HealthCheck func(context.Context) error
	Auth        *auth.Service
	AdminStore  *admin.Store
}

func NewRouter(options Options) http.Handler {
	if options.Now == nil {
		options.Now = time.Now
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandlerWithCheck(options.Now, options.HealthCheck))
	if options.Auth != nil {
		registerAuthRoutes(mux, options.Auth)
		if options.AdminStore != nil {
			registerAdminAPIRoutes(mux, options.Auth, options.AdminStore)
			registerPageRoutes(mux, options.Auth, options.AdminStore)
		} else {
			registerPageRoutes(mux, options.Auth, nil)
		}
	}
	return withJSONHeaders(mux)
}

func withJSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}
