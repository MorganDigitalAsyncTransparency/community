#!/bin/sh
if [ -f frontend/package.json ] && [ -d frontend/node_modules ]; then
  (cd frontend && npx tsc --noEmit)
else
  echo "No frontend dependencies installed, skipping type check."
fi
