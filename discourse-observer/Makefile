.PHONY: setup lint lint-go lint-md test build up down start restart open

## Setup — install dependencies and configure git hooks
setup:
	npm install
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	git config core.hooksPath .githooks

## Lint — run all linters
lint: lint-go lint-md

## Go linting (skips if no Go files exist)
lint-go:
	@sh scripts/lint-go.sh

## Markdown linting
lint-md:
	npx markdownlint "**/*.md" --ignore node_modules --ignore web

## Tests (skips if no Go files exist)
test:
	@sh scripts/test-go.sh

## Docker — build and run containers
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

## Start — one-command onboarding: configure, build, launch, open browser
start: .env build up open

.env:
	cp .env.example .env
	@echo "Created .env from .env.example — edit it with your Discourse credentials."

## Restart — rebuild and relaunch after code changes
restart: down build up open

## Open the dashboard in the default browser
open:
	@command -v xdg-open >/dev/null 2>&1 && xdg-open http://localhost:3000 \
	|| command -v open >/dev/null 2>&1 && open http://localhost:3000 \
	|| start http://localhost:3000 2>/dev/null \
	|| echo "Open http://localhost:3000 in your browser"
