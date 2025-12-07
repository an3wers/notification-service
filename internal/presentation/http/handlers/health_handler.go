package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)

	defer cancel()

	response := HealthResponse{
		Status:   "healthy",
		Services: make(map[string]string),
	}

	// Check database
	if err := h.db.Ping(ctx); err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "unhealthy: " + err.Error()
	} else {
		response.Services["database"] = "healthy"
	}

	status := http.StatusOK

	if response.Status == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
