#!/bin/sh
if [ -f frontend/package.json ] && [ -d frontend/node_modules ]; then
  (cd frontend && npm test)
else
  echo "No frontend dependencies installed, skipping frontend tests."
fi
