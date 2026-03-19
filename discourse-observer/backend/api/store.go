// Spec: specs/api/api-contract.md (AC-8, AC-9, AC-10)
// Tests: backend/api/contract_test.go
package api

import (
	"context"

	"github.com/code-community/discourse-observer/backend/model"
)

// TopicReader loads topics from a persistent store.
// Defined here so the api package depends on an abstraction,
// not on the storage implementation directly.
type TopicReader interface {
	QueryTopics(ctx context.Context, opts model.QueryOpts) ([]model.Topic, error)
}
