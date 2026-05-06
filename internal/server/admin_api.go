package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/auth"
	"github.com/liangzh77/keychain/internal/web"
)

func registerAdminAPIRoutes(mux *http.ServeMux, authService *auth.Service, store *admin.Store) {
	mux.HandleFunc("GET /api/providers", requireAdmin(authService, listProvidersHandler(store)))
	mux.HandleFunc("POST /api/providers", requireAdmin(authService, createProviderHandler(store)))
	mux.HandleFunc("GET /api/models", requireAdmin(authService, listModelsHandler(store)))
	mux.HandleFunc("POST /api/models", requireAdmin(authService, createModelHandler(store)))
	mux.HandleFunc("GET /api/keys", requireAdmin(authService, listKeysHandler(store)))
	mux.HandleFunc("POST /api/keys", requireAdmin(authService, createKeyHandler(store)))
}

func listProvidersHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providers, err := store.ListProviders(r.Context())
		if err != nil {
			web.WriteError(w, http.StatusInternalServerError, "LIST_PROVIDERS_FAILED", "Failed to list providers", nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, providers)
	}
}

func createProviderHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input admin.CreateProviderInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		provider, err := store.CreateProvider(r.Context(), input)
		if err != nil {
			web.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusCreated, provider)
	}
}

func listModelsHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := store.ListModels(r.Context(), r.URL.Query().Get("providerId"), r.URL.Query().Get("providerCode"))
		if err != nil {
			web.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, models)
	}
}

func createModelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input admin.CreateModelInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		model, err := store.CreateModel(r.Context(), input)
		if err != nil {
			web.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusCreated, model)
	}
}

func listKeysHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := store.ListAPIKeys(r.Context(), r.URL.Query().Get("providerId"))
		if err != nil {
			web.WriteError(w, http.StatusInternalServerError, "LIST_KEYS_FAILED", "Failed to list keys", nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, keys)
	}
}

func createKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input admin.CreateAPIKeyInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		key, err := store.CreateAPIKey(r.Context(), input)
		if err != nil {
			web.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusCreated, key)
	}
}

func parseOptionalInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
