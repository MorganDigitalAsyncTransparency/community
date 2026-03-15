#!/bin/sh
if find backend/ -name '*.go' 2>/dev/null | grep -q .; then
  golangci-lint run ./backend/...
else
  echo "No Go files, skipping Go lint."
fi
