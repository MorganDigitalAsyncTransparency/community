// Spec: specs/api/api-contract.md (AC-6, AC-7)
// Tests: backend/api/contract_test.go
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("response write failed: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("error response write failed: %v", err)
	}
}
