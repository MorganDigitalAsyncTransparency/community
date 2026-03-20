// Seed populates the SQLite database with mock data by running the
// data pipeline against the built-in mock Discourse server.
// Used for development when no real Discourse forum is available.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath := os.Getenv("OBSERVER_DB")
	if dbPath == "" {
		dbPath = "data/analytics.db"
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = store.Close() }()

	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	obs := observer.New(client, store, srv.URL)

	ctx := context.Background()
	if _, err := obs.Run(ctx); err != nil {
		return fmt.Errorf("pipeline run: %w", err)
	}

	topics, err := store.LoadTopics(ctx)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	fmt.Printf("Seeded %d topics into %s\n", len(topics), dbPath)
	return nil
}
