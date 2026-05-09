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
	Name      string `json:"name"`
	IsEnabled *bool  `json:"isEnabled"`
}

type runtimeUserResponse struct {
	ID             string `json:"id"`
	ChannelID      string `json:"channelId"`
	ExternalUserID string `json:"externalUserId"`
	Name           string `json:"name"`
	IsEnabled      bool   `json:"isEnabled"`
}

type hostedUserRegisterRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type hostedUserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type hostedUserResetPasswordRequest struct {
	Password string `json:"password"`
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
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
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
	mux.HandleFunc("PUT /api/runtime/channels/{channelId}/external-users/{externalUserId}", requireRuntime(runtimeUpsertExternalUserHandler(store)))
	mux.HandleFunc("DELETE /api/runtime/channels/{channelId}/external-users/{externalUserId}", requireRuntime(runtimeDeleteExternalUserHandler(store)))
	mux.HandleFunc("POST /api/runtime/channels/{channelId}/hosted-users/register", requireRuntime(runtimeRegisterHostedUserHandler(store)))
	mux.HandleFunc("POST /api/runtime/channels/{channelId}/hosted-users/login", requireRuntime(runtimeLoginHostedUserHandler(store)))
	mux.HandleFunc("POST /api/runtime/channels/{channelId}/hosted-users/{userId}/reset-password", requireRuntime(runtimeResetHostedUserPasswordHandler(store)))
	mux.HandleFunc("DELETE /api/runtime/channels/{channelId}/hosted-users/{userId}", requireRuntime(runtimeDeleteHostedUserHandler(store)))
	mux.HandleFunc("GET /api/runtime/users/{id}/permissions", requireRuntime(runtimeUserPermissionsHandler(store)))
	mux.HandleFunc("GET /api/runtime/providers", requireRuntime(runtimeProvidersHandler(store)))
	mux.HandleFunc("GET /api/runtime/models", requireRuntime(runtimeModelsHandler(store)))
	mux.HandleFunc("POST /api/runtime/dispatches", requireRuntime(runtimeDispatchKeyHandler(store)))
	mux.HandleFunc("POST /api/runtime/dispatches/{id}/failure", requireRuntime(runtimeKeyFailureHandler(store)))
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

func runtimeUpsertExternalUserHandler(store *admin.Store) http.HandlerFunc {
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
		user, err := store.UpsertRuntimeExternalUser(r.Context(), admin.UpsertRuntimeExternalUserInput{
			ChannelID:      r.PathValue("channelId"),
			ExternalUserID: r.PathValue("externalUserId"),
			Name:           request.Name,
			IsEnabled:      isEnabled,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, runtimeUserResponse{
			ID:             user.ID,
			ChannelID:      user.ChannelID,
			ExternalUserID: user.ExternalUserID,
			Name:           user.DisplayName,
			IsEnabled:      user.IsEnabled,
		})
	}
}

func runtimeDeleteExternalUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.DeleteRuntimeExternalUser(r.Context(), r.PathValue("channelId"), r.PathValue("externalUserId")); err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "DELETE_USER_FAILED", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

func runtimeRegisterHostedUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request hostedUserRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		user, err := store.RegisterRuntimeHostedUser(r.Context(), admin.RegisterRuntimeHostedUserInput{
			ChannelID: r.PathValue("channelId"),
			Username:  request.Username,
			Name:      request.Name,
			Password:  request.Password,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "REGISTER_HOSTED_USER_FAILED", err.Error(), nil)
			return
		}
		writeRuntimeUser(w, user)
	}
}

func runtimeLoginHostedUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request hostedUserLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		user, err := store.LoginRuntimeHostedUser(r.Context(), admin.LoginRuntimeHostedUserInput{
			ChannelID: r.PathValue("channelId"),
			Username:  request.Username,
			Password:  request.Password,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "LOGIN_HOSTED_USER_FAILED", err.Error(), nil)
			return
		}
		writeRuntimeUser(w, user)
	}
}

func runtimeResetHostedUserPasswordHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request hostedUserResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			web.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
			return
		}
		user, err := store.ResetRuntimeHostedUserPassword(r.Context(), admin.ResetRuntimeHostedUserPasswordInput{
			ChannelID: r.PathValue("channelId"),
			UserID:    r.PathValue("userId"),
			Password:  request.Password,
		})
		if err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "RESET_HOSTED_USER_PASSWORD_FAILED", err.Error(), nil)
			return
		}
		writeRuntimeUser(w, user)
	}
}

func runtimeDeleteHostedUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.DeleteRuntimeHostedUser(r.Context(), r.PathValue("channelId"), r.PathValue("userId")); err != nil {
			web.WriteError(w, runtimeErrorStatus(err), "DELETE_HOSTED_USER_FAILED", err.Error(), nil)
			return
		}
		web.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

func writeRuntimeUser(w http.ResponseWriter, user admin.User) {
	web.WriteJSON(w, http.StatusOK, runtimeUserResponse{
		ID:             user.ID,
		ChannelID:      user.ChannelID,
		ExternalUserID: user.ExternalUserID,
		Name:           user.DisplayName,
		IsEnabled:      user.IsEnabled,
	})
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
		result, err := store.ReportRuntimeKeyFailure(r.Context(), r.PathValue("id"), request.ErrorCode, request.ErrorMessage)
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
	if strings.Contains(message, "invalid credentials") {
		return http.StatusUnauthorized
	}
	if strings.Contains(message, "not found") {
		return http.StatusNotFound
	}
	if strings.Contains(message, "already exists") {
		return http.StatusConflict
	}
	if strings.Contains(message, "permission denied") {
		return http.StatusForbidden
	}
	return http.StatusUnprocessableEntity
}
