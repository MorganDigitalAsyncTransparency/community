.PHONY: setup lint lint-go lint-md test

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
	npx markdownlint "**/*.md" --ignore node_modules

## Tests (skips if no Go files exist)
test:
	@sh scripts/test-go.sh
