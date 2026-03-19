package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/code-community/discourse-observer/backend/api"
	"github.com/code-community/discourse-observer/backend/domain"
	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/storage"
)

func main() {
	tagConfig, err := loadTagConfig("config/tagConfig.json")
	if err != nil {
		log.Fatalf("failed to load tag config: %v", err)
	}
	buckets, err := loadDistributionBuckets("config/distributionBuckets.json")
	if err != nil {
		log.Fatalf("failed to load distribution buckets: %v", err)
	}

	dbPath := os.Getenv("OBSERVER_DB")
	if dbPath == "" {
		dbPath = "data/analytics.db"
	}
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = store.Close() }()

	srv := &api.Server{
		Store:          store,
		TagConfig:      tagConfig,
		ResolvedTags:   domain.ResolveAllTags(&tagConfig),
		BucketCeilings: buckets.BucketCeilingsHours,
		Version:        "0.1.0",
		Now:            func() time.Time { return time.Now().UTC() },
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	srv.RegisterRoutes(mux)

	log.Println("backend listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("server stopped: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("health response write failed: %v", err)
	}
}

func loadTagConfig(path string) (model.TagConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.TagConfig{}, err
	}
	var cfg model.TagConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return model.TagConfig{}, err
	}
	return cfg, nil
}

func loadDistributionBuckets(path string) (model.DistributionBuckets, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.DistributionBuckets{}, err
	}
	var b model.DistributionBuckets
	if err := json.Unmarshal(data, &b); err != nil {
		return model.DistributionBuckets{}, err
	}
	return b, nil
}
