// Spec: specs/observer/mock-server-service.md (MS-2)
// Tests: go build ./backend/cmd/mockserver (compile check)
//
// Standalone HTTP server for the mock Discourse API.
// Serves realistic topic and category data from built-in fixtures.
// Used as a docker-compose service so the sync pipeline works in dev mode.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
)

func main() {
	handler := mockserver.Handler()
	if v := os.Getenv("MOCK_PAGE_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			log.Fatalf("invalid MOCK_PAGE_SIZE: %s", v)
		}
		handler = mockserver.HandlerWithPageSize(n)
	}

	srv := &http.Server{
		Addr:    ":9920",
		Handler: handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Println("mock discourse server listening on :9920")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		stop()
		log.Fatalf("server stopped: %v", err)
	}
	stop()
}
