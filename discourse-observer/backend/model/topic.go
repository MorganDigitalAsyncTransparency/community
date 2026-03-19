// Spec: specs/api/api-contract.md
// Tests: backend/domain/*_test.go, backend/api/contract_test.go
package model

import "time"

// Topic represents a normalized forum topic used across all domain calculations.
type Topic struct {
	ID             int
	Title          string
	CreatedAt      time.Time
	Tags           []string
	CategoryName   string
	ReplyCount     int
	FirstReplyAt   *time.Time
	ResolvedAt     *time.Time
	Outcome        string // "solved", "self-closed", or "" (open)
	LastActivityAt *time.Time
	TopicURL       string
}
