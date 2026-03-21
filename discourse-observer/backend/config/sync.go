// Spec: specs/observer/scheduler.md (SC-1, SC-2)
// Tests: backend/scheduler/scheduler_acceptance_test.go
package config

import (
	"os"
	"strconv"
	"time"
)

// SyncConfig holds scheduling parameters for the sync lifecycle.
type SyncConfig struct {
	InitialDelay         time.Duration
	DeltaDelay           time.Duration
	Interval             time.Duration
	LowActivityThreshold int
	MaxRetries           int
	JitterMax            time.Duration
}

// LoadSyncConfig reads sync configuration from environment variables,
// returning defaults for any unset variable.
func LoadSyncConfig() SyncConfig {
	return SyncConfig{
		InitialDelay:         durationEnv("SYNC_INITIAL_DELAY_SECONDS", 20*time.Second),
		DeltaDelay:           durationEnv("SYNC_DELTA_DELAY_SECONDS", 2*time.Second),
		Interval:             durationEnv("SYNC_INTERVAL", 15*time.Minute),
		LowActivityThreshold: intEnv("SYNC_LOW_ACTIVITY_THRESHOLD", 3),
		MaxRetries:           intEnv("SYNC_MAX_RETRIES", 3),
		JitterMax:            durationEnv("SYNC_JITTER_SECONDS", 60*time.Second),
	}
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	s, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return time.Duration(s) * time.Second
}

func intEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
