package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/liangzh77/keychain/internal/auth"
	"github.com/liangzh77/keychain/internal/web"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	Username string `json:"username"`
}

func registerAuthRoutes(mux *http.ServeMux, authService *auth.Service) {
	mux.HandleFunc("POST /api/auth/login", apiLoginHandler(authService))
	mux.HandleFunc("POST /api/auth/logout", apiLogoutHandler(authService))
	mux.HandleFunc("GET /api/auth/me", requireAdmin(authService, apiMeHandler(authService)))
}

func apiLoginHandler(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request loginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		if !authService.Authenticate(request.Username, request.Password) {
			web.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials", nil)
			return
		}

		token, session, err := authService.CreateSession(r.Context())
		if err != nil {
			web.WriteError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", "Failed to create session", nil)
			return
		}
		setSessionCookie(w, r, token, session.ExpiresAt)
		web.WriteJSON(w, http.StatusOK, meResponse{Username: session.Username})
	}
}

func apiLogoutHandler(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(auth.CookieName); err == nil {
			_ = authService.DeleteSession(r.Context(), cookie.Value)
		}
		clearSessionCookie(w, r)
		web.WriteJSON(w, http.StatusOK, map[string]bool{"loggedOut": true})
	}
}

func apiMeHandler(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, meResponse{Username: authService.AdminUsername()})
	}
}

func requireAdmin(authService *auth.Service, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			web.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
			return
		}
		if _, ok, err := authService.GetSession(r.Context(), cookie.Value); err != nil {
			web.WriteError(w, http.StatusInternalServerError, "SESSION_CHECK_FAILED", "Failed to check session", nil)
			return
		} else if !ok {
			clearSessionCookie(w, r)
			web.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
			return
		}
		next(w, r)
	}
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}
