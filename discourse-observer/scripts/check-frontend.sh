#!/bin/sh
if [ -f web/package.json ] && [ -d web/node_modules ]; then
  (cd web && npx tsc --noEmit)
else
  echo "No frontend dependencies installed, skipping type check."
fi
