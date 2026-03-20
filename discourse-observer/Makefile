.PHONY: check setup lint lint-go lint-md lint-frontend test verify docs build up down start restart open seed maybe-seed

## Check — verify prerequisites are installed
check:
	@sh scripts/setup.sh

## Setup — install dependencies and configure git hooks
setup:
	npm install
	cd frontend && npm install
	pip install -r docs/requirements.txt
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	git config core.hooksPath .githooks

## Lint — run all linters
lint: lint-go lint-md lint-frontend

## Go linting (skips if no Go files exist)
lint-go:
	@sh scripts/lint-go.sh

## Markdown linting
lint-md:
	npx markdownlint "**/*.md" --ignore "**/node_modules/**" --ignore "**/test-reports/**"

## Frontend linting — ESLint and Stylelint (skips if no dependencies installed)
lint-frontend:
	@sh scripts/lint-frontend.sh

## Tests — run all tests
test:
	@sh scripts/test-go.sh
	@sh scripts/test-frontend.sh

## Verify — run all linters and tests
verify: lint test

## Docs — serve documentation locally
docs:
	python -m mkdocs serve -f docs/mkdocs.yml --watch docs --watch specs --watch ARCHITECTURE.md --watch CONTRIBUTING.md --watch README.md

## Docker — build and run containers
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down --timeout 10 --remove-orphans

## Start — one-command onboarding: setup, configure, build, launch, open browser
## Seeds mock data automatically when DISCOURSE_API_TOKEN is empty in .env.
start: setup .env config/tagConfig.json verify maybe-seed build up open

.env:
	cp .env.example .env
	@echo "Created .env from .env.example — edit it with your Discourse credentials."

config/tagConfig.json:
	cp config/tagConfig.example.json config/tagConfig.json
	@echo "Created config/tagConfig.json from example — edit it with your tag and SLO configuration."

## Seed — populate SQLite with mock data for development
seed:
	@mkdir -p data
	go run ./backend/cmd/seed

## Maybe-seed — seed only when no real Discourse credentials are configured
maybe-seed:
	@TOKEN=$$(grep -s '^DISCOURSE_API_TOKEN=' .env | cut -d= -f2-); \
	if [ -z "$$TOKEN" ]; then \
		echo "No DISCOURSE_API_TOKEN in .env — seeding mock data..."; \
		mkdir -p data; \
		go run ./backend/cmd/seed; \
	else \
		echo "DISCOURSE_API_TOKEN found — skipping seed."; \
	fi

## Restart — rebuild and relaunch after code changes
restart: down verify build up open

## Open the dashboard in the default browser
open:
	@command -v xdg-open >/dev/null 2>&1 && xdg-open http://localhost:3000 \
	|| command -v open >/dev/null 2>&1 && open http://localhost:3000 \
	|| start http://localhost:3000 2>/dev/null \
	|| true
	@echo "Dashboard running at http://localhost:3000 (opening in browser...)"
	@echo "API health check at http://localhost:3000/api/v1/status"
