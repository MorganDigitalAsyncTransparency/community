// Spec: specs/observer/mock-server-service.md
// Tests: backend/api/contract_test.go
package api

import "net/http"

func (s *Server) handleSyncLog(w http.ResponseWriter, _ *http.Request) {
	if s.SyncStatus == nil {
		respondJSON(w, []any{})
		return
	}

	entries := s.SyncStatus.GetLog()
	type jsonEntry struct {
		Timestamp string  `json:"timestamp"`
		Mode      string  `json:"mode"`
		Pages     int     `json:"pages"`
		Topics    int     `json:"topics"`
		Duration  float64 `json:"durationSeconds"`
	}

	out := make([]jsonEntry, len(entries))
	for i, e := range entries {
		out[i] = jsonEntry{
			Timestamp: e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			Mode:      e.Mode,
			Pages:     e.Pages,
			Topics:    e.Topics,
			Duration:  e.Duration.Seconds(),
		}
	}
	respondJSON(w, out)
}
