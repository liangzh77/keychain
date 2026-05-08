package server

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/web"
)

type runtimeUserRequest struct {
	ChannelID string `json:"channelId"`
	Name      string `json:"name"`
	IsEnabled *bool  `json:"isEnabled"`
}

type runtimeUserResponse struct {
	ID        string `json:"id"`
	ChannelID string `json:"channelId"`
	Name      string `json:"name"`
	IsEnabled bool   `json:"isEnabled"`
}

type runtimePermissionsResponse struct {
	UserID      string                      `json:"userId"`
	Permissions []admin.EffectivePermission `json:"permissions"`
}

type dispatchKeyRequest struct {
	ChannelID  string `json:"channelId"`
	UserID     string `json:"userId"`
	ProviderID string `json:"providerId"`
	ModelID    string `json:"modelId"`
}

type dispatchKeyResponse struct {
	DispatchLogID string `json:"dispatchLogId"`
	ProviderName  string `json:"providerName"`
	ModelName     string `json:"modelName"`
	KeyID         string `json:"keyId"`
	KeyAlias      string `json:"keyAlias"`
	Key           string `json:"key"`
}

type keyFailureRequest struct {
	DispatchLogID string `json:"dispatchLogId"`
	ErrorCode     string `json:"errorCode"`
	ErrorMessage  string `json:"errorMessage"`
}

type keyFailureResponse struct {
	Reported    bool   `json:"reported"`
	KeyID       string `json:"keyId"`
	KeyAlias    string `json:"keyAlias"`
	IsAvailable bool   `json:"isAvailable"`
}

func registerRuntimeAPIRoutes(mux *http.ServeMux, store *admin.Store, token string) {
	requireRuntime := func(next http.HandlerFunc) http.HandlerFunc {
		return requireRuntimeToken(token, next)
	}
	mux.HandleFunc("POST /api/runtime/users", requireRuntime(runtimeUpsertUserHandler(store)))
	mux.HandleFunc("DELETE /api/runtime/users/{id}", requireRuntime(runtimeDeleteUserHandler(store)))
	mux.HandleFunc("GET /api/runtime/users/{id}/permissions", requireRuntime(runtimeUserPermissionsHandler(store)))
	mux.HandleFunc("GET /api/runtime/providers", requireRuntime(runtimeProvidersHandler(store)))
	mux.HandleFunc("GET /api/runtime/models", requireRuntime(runtimeModelsHandler(store)))
	mux.HandleFunc("POST /api/runtime/dispatch-key", requireRuntime(runtimeDispatchKeyHandler(store)))
	mux.HandleFunc("POST /api/runtime/key-failures", requireRuntime(runtimeKeyFailureHandler(store)))
}

func requireRuntimeToken(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(token) == "" {
			web.WriteError(w, http.StatusServiceUnavailable, "RUNTIME_API_DISABLED", "Runtime API is not configured", nil)
			return
		}
		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, prefix) {
			web.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Runtime API token required", nil)
			return
		}
		presented := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
		if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
			web.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Runtime API token required", nil)
			return
		}
		next(w, r)
	}
}

func runtimeUpsertUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request runtimeUserRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		isEnabled := true
		if request.IsEnabled != nil {
			isEnabled = *request.IsEnabled
		}
		user, err := store.UpsertRuntimeUser(r.Context(), admin.UpsertRuntimeUserInput{
			ChannelID: request.ChannelID,
			Name:      request.Name,
			IsEnabled: isEnabled,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, runtimeUserResponse{
			ID:        user.ID,
			ChannelID: user.ChannelID,
			Name:      user.DisplayName,
			IsEnabled: user.IsEnabled,
		})
	}
}

func runtimeDeleteUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if err := store.DeleteUser(r.Context(), userID); err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "DELETE_USER_FAILED", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

func runtimeUserPermissionsHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		permissions, err := store.ListEffectiveUserPermissions(r.Context(), userID)
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, runtimePermissionsResponse{UserID: userID, Permissions: permissions})
	}
}

func runtimeProvidersHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providers, err := store.ListRuntimeProviders(r.Context())
		if err != nil {
			web.WriteError(w, http.StatusInternalServerError, "LIST_PROVIDERS_FAILED", "Failed to list providers", nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, providers)
	}
}

func runtimeModelsHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := store.ListRuntimeModels(r.Context(), r.URL.Query().Get("providerId"))
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, models)
	}
}

func runtimeDispatchKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request dispatchKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		result, err := store.DispatchRuntimeKey(r.Context(), admin.DispatchKeyInput{
			ChannelID:  request.ChannelID,
			UserID:     request.UserID,
			ProviderID: request.ProviderID,
			ModelID:    request.ModelID,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "DISPATCH_KEY_FAILED", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, dispatchKeyResponse{
			DispatchLogID: result.DispatchLogID,
			ProviderName:  result.ProviderName,
			ModelName:     result.ModelName,
			KeyID:         result.KeyID,
			KeyAlias:      result.KeyAlias,
			Key:           result.Key,
		})
	}
}

func runtimeKeyFailureHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request keyFailureRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		result, err := store.ReportRuntimeKeyFailure(r.Context(), request.DispatchLogID, request.ErrorCode, request.ErrorMessage)
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "KEY_FAILURE_REPORT_FAILED", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, keyFailureResponse{
			Reported:    true,
			KeyID:       result.KeyID,
			KeyAlias:    result.KeyAlias,
			IsAvailable: result.IsAvailable,
		})
	}
}

func runtimeErrorStatus(err error) int {
	message := err.Error()
	if strings.Contains(message, "not found") {
		return http.StatusNotFound
	}
	if strings.Contains(message, "permission denied") {
		return http.StatusForbidden
	}
	return http.StatusUnprocessableEntity
}
