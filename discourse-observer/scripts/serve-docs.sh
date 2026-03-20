#!/bin/sh
# Serve docs with poll-based rebuild.
#
# mkdocs serve live reload fails on Windows because docs_dir: ..
# makes it watch the entire project tree including node_modules.
# This script works around that by polling for changes and
# rebuilding with mkdocs build --dirty.
#
# The browser does not auto-refresh — reload manually after edits.

set -e
CONFIG="docs/mkdocs.yml"
PORT=8000
SITE_DIR="../site"
POLL=2

echo "Building documentation..."
python -m mkdocs build -f "$CONFIG" -q

echo "Serving on http://127.0.0.1:$PORT/"
echo "Watching docs/ and specs/ for changes (poll every ${POLL}s)."
echo "Browser does not auto-refresh — reload manually after edits."
echo ""

# Start HTTP server in background
python -m http.server "$PORT" -d "$SITE_DIR" -b 127.0.0.1 &
SERVER_PID=$!
cleanup() {
    kill $SERVER_PID 2>/dev/null
    rm -rf "$SITE_DIR"
    echo "[docs] Cleaned up $SITE_DIR"
}
trap cleanup EXIT INT TERM

# Poll for changes and rebuild
STAMP=$(date +%s)
while kill -0 "$SERVER_PID" 2>/dev/null; do
    sleep "$POLL"
    # Find .md and .yml files newer than last check
    CHANGED=$(find docs specs ARCHITECTURE.md CONTRIBUTING.md README.md \
        -name '*.md' -o -name '*.yml' 2>/dev/null | while read -r f; do
        ts=$(date -r "$f" +%s 2>/dev/null || stat -c %Y "$f" 2>/dev/null || echo 0)
        if [ "$ts" -gt "$STAMP" ]; then echo "$f"; fi
    done)
    if [ -n "$CHANGED" ]; then
        echo "[docs] Changed: $CHANGED"
        python -m mkdocs build -f "$CONFIG" --dirty -q 2>&1 || true
        STAMP=$(date +%s)
        echo "[docs] Rebuilt — refresh your browser."
    fi
done
