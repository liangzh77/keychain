package server

import (
	"context"
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
	return healthHandlerWithCheck(now, nil)
}

func healthHandlerWithCheck(now func() time.Time, healthCheck func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		databaseStatus := "not_configured"
		if healthCheck != nil {
			databaseStatus = "ok"
			if err := healthCheck(r.Context()); err != nil {
				databaseStatus = "error"
			}
		}
		web.WriteJSON(w, http.StatusOK, healthResponse{
			Status:   "ok",
			Database: databaseStatus,
			Time:     now().UTC().Format(time.RFC3339),
		})
	}
}
