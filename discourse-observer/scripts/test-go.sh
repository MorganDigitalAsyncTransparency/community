#!/bin/sh
if find backend/ -name '*.go' 2>/dev/null | grep -q .; then
  go test ./backend/...
else
  echo "No Go files, skipping tests."
fi
