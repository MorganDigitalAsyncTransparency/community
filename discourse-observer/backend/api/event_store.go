// Spec: specs/api/triage-time.md (TT-8)
// Tests: backend/api/triage-time_contract_test.go
package api

import (
	"context"

	"github.com/code-community/discourse-observer/backend/model"
)

// EventReader loads topic events from a persistent store.
// Defined here so the api package depends on an abstraction,
// not on the storage implementation directly.
type EventReader interface {
	LoadAllTopicEvents(ctx context.Context) (map[int][]model.TopicEvent, error)
}
