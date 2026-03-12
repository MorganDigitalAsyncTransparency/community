#!/bin/sh
if find . -name '*.go' -not -path './vendor/*' | grep -q .; then
  golangci-lint run ./...
else
  echo "No Go files, skipping Go lint."
fi
