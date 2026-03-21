// Spec: specs/api/api-contract.md (AC-33)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"
	"time"
)

func (s *Server) handleSyncLog(w http.ResponseWriter, _ *http.Request) {
	if s.SyncStatus == nil {
		respondJSON(w, map[string]any{"progress": nil, "entries": []any{}})
		return
	}

	type jsonProgress struct {
		Mode     string  `json:"mode"`
		Pages    int     `json:"pages"`
		Topics   int     `json:"topics"`
		ElapsedS float64 `json:"elapsedSeconds"`
	}

	type jsonEntry struct {
		Timestamp string  `json:"timestamp"`
		Mode      string  `json:"mode"`
		Pages     int     `json:"pages"`
		Topics    int     `json:"topics"`
		Duration  float64 `json:"durationSeconds"`
	}

	var prog *jsonProgress
	if p := s.SyncStatus.GetProgress(); p != nil {
		prog = &jsonProgress{
			Mode:     p.Mode,
			Pages:    p.Pages,
			Topics:   p.Topics,
			ElapsedS: time.Since(p.StartedAt).Seconds(),
		}
	}

	entries := s.SyncStatus.GetLog()
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

	respondJSON(w, map[string]any{"progress": prog, "entries": out})
}
