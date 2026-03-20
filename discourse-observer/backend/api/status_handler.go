// Spec: specs/api/api-contract.md (AC-28)
// Tests: backend/api/contract_test.go
package api

import "net/http"

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]any{
		"version": s.Version,
	}

	if s.SyncStatus == nil {
		resp["syncState"] = "disabled"
		resp["lastSyncedAt"] = nil
		respondJSON(w, resp)
		return
	}

	syncedAt := s.SyncStatus.GetLastSyncedAt()
	var lastSynced *string
	if syncedAt != nil {
		t := syncedAt.UTC().Format("2006-01-02T15:04:05Z")
		lastSynced = &t
	}

	resp["lastSyncedAt"] = lastSynced
	resp["syncState"] = s.SyncStatus.GetState()
	resp["lastSyncDuration"] = s.SyncStatus.GetLastDuration().Seconds()
	resp["lastSyncTopics"] = s.SyncStatus.GetLastTopics()

	respondJSON(w, resp)
}
