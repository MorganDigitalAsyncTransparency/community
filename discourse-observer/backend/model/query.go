package model

import "time"

// QueryOpts holds resolved filter parameters for topic queries.
// Time bounds are concrete timestamps (period is resolved by the caller).
type QueryOpts struct {
	From *time.Time
	To   *time.Time
	Tag  string
}
