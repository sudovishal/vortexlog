package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sudovishal/vortexlog/internal/database"
)

type LogPayload struct {
	ServiceName string `json:"service_name"`
	LogLevel    string `json:"log_level"`
	Message     string `json:"message"`
	// Timestamp   time.Time `json:"timestamp"`
}

func sendSampleLog(ctx context.Context, queries *database.Queries, payload LogPayload) error {
	log, err := queries.CreateLog(ctx, database.CreateLogParams{
		ServiceName: payload.ServiceName,
		LogLevel:    payload.LogLevel,
		Message:     payload.Message,
	})
	if err != nil {
		return fmt.Errorf("failed to insert sample log: %w", err)
	}

	fmt.Printf("Successfully inserted log with ID: %d\n", log.ID)
	return nil
}

func HandleLogs(logChan chan<- LogPayload) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all responses
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		var payload LogPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondWithError(w, 400, "Invalid JSON payload", err)
			return
		}

		logChan <- payload

		// Directly call DB save function
		// ctx := r.Context()
		// if err := sendSampleLog(ctx, queries, payload); err != nil {
		// 	respondWithError(w, http.StatusInternalServerError, "Failed to save log", err)
		// 	return
		// }

		respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
	}
}
