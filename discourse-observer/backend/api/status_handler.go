// Spec: specs/api/api-contract.md (AC-28)
// Tests: backend/api/contract_test.go
package api

import "net/http"

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	var lastSynced *string
	if s.LastSyncedAt != nil {
		t := s.LastSyncedAt.UTC().Format("2006-01-02T15:04:05Z")
		lastSynced = &t
	}
	respondJSON(w, map[string]any{
		"lastSyncedAt": lastSynced,
		"version":      s.Version,
	})
}
