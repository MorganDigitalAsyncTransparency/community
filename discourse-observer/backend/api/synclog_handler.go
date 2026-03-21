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
		Mode        string  `json:"mode"`
		Topics      int     `json:"topics"`
		TotalTopics int     `json:"totalTopics"`
		ElapsedS    float64 `json:"elapsedSeconds"`
		EtaSeconds  float64 `json:"etaSeconds"`
	}

	type jsonEntry struct {
		Timestamp  string  `json:"timestamp"`
		Mode       string  `json:"mode"`
		Topics     int     `json:"topics"`
		Duration   float64 `json:"durationSeconds"`
		HasChanges bool    `json:"hasChanges"`
		Error      string  `json:"error"`
	}

	var prog *jsonProgress
	if p := s.SyncStatus.GetProgress(); p != nil {
		elapsed := time.Since(p.StartedAt).Seconds()
		var eta float64
		if p.TotalTopics > 0 && p.Topics > 0 {
			rate := elapsed / float64(p.Topics)
			remaining := float64(p.TotalTopics-p.Topics) * rate
			if remaining > 0 {
				eta = remaining
			}
		}
		prog = &jsonProgress{
			Mode:        p.Mode,
			Topics:      p.Topics,
			TotalTopics: p.TotalTopics,
			ElapsedS:    elapsed,
			EtaSeconds:  eta,
		}
	}

	entries := s.SyncStatus.GetLog()
	out := make([]jsonEntry, len(entries))
	for i, e := range entries {
		out[i] = jsonEntry{
			Timestamp:  e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			Mode:       e.Mode,
			Topics:     e.Topics,
			Duration:   e.Duration.Seconds(),
			HasChanges: e.HasChanges,
			Error:      e.Error,
		}
	}

	respondJSON(w, map[string]any{"progress": prog, "entries": out})
}
