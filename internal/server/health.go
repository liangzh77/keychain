package server

import (
	"net/http"
	"time"

	"github.com/liangzh77/keychain/internal/web"
)

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Time     string `json:"time"`
}

func healthHandler(now func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, healthResponse{
			Status:   "ok",
			Database: "not_configured",
			Time:     now().UTC().Format(time.RFC3339),
		})
	}
}
