package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// ---- utils ----
// respondWithError writes an error response as JSON
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON writes the response as JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ---- handlers ----
// BaseHandler contains common handler dependencies
type BaseHandler struct {
	DB *pgx.Conn
}
