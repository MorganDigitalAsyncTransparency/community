#!/bin/sh
if find . -name '*.go' -not -path './vendor/*' | grep -q .; then
  go test ./...
else
  echo "No Go files, skipping tests."
fi
