package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/code-community/discourse-observer/backend/api"
	"github.com/code-community/discourse-observer/backend/config"
	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/domain"
	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/scheduler"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	sched := startSyncIfConfigured(ctx, store, srv)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	srv.RegisterRoutes(mux)

	httpSrv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	log.Println("backend listening on :8080")
	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
	}

	if sched != nil {
		// Context is already canceled; scheduler will finish its in-progress sync and return.
		log.Println("waiting for scheduler shutdown...")
	}
}

// startSyncIfConfigured creates and starts the scheduler if Discourse
// credentials are present. Returns nil if sync is disabled (dev mode).
func startSyncIfConfigured(ctx context.Context, store *storage.SQLiteStore, srv *api.Server) *scheduler.Scheduler {
	discourseURL := os.Getenv("DISCOURSE_URL")
	if discourseURL == "" {
		log.Println("DISCOURSE_URL not set — sync disabled (dev mode)")
		return nil
	}

	apiKey := os.Getenv("DISCOURSE_API_KEY")
	apiUser := os.Getenv("DISCOURSE_API_USERNAME")
	syncCfg := config.LoadSyncConfig()

	pageCfg := discourse.PageConfig{
		Delay:      syncCfg.InitialDelay,
		MaxRetries: syncCfg.MaxRetries,
		RetryDelay: 60 * time.Second,
	}
	client := discourse.NewClient(discourseURL, apiKey, apiUser, discourse.WithPageConfig(pageCfg))
	obs := observer.New(client, store, discourseURL)

	sched := scheduler.New(obs, syncCfg)
	srv.SyncStatus = sched.Status()

	go sched.Start(ctx)
	log.Println("sync scheduler started")

	return sched
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
