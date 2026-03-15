#!/bin/sh
set -e

# Colors (disabled when output is not a terminal)
if [ -t 1 ]; then
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  YELLOW='\033[0;33m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  GREEN='' RED='' YELLOW='' BOLD='' RESET=''
fi

pass() { printf "  ${GREEN}✓${RESET} %s\n" "$1"; }
fail() { printf "  ${RED}✗${RESET} %s — %s\n" "$1" "$2"; }
warn() { printf "  ${YELLOW}!${RESET} %s — %s\n" "$1" "$2"; }

required_ok=true
dev_ok=true

# --- Detect OS for install hints ---

detect_os() {
  case "$(uname -s)" in
    Linux*)
      if [ -f /etc/debian_version ]; then echo "debian"
      elif [ -f /etc/fedora-release ]; then echo "fedora"
      else echo "linux"
      fi ;;
    Darwin*) echo "macos" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "unknown" ;;
  esac
}

OS=$(detect_os)

make_hint() {
  case "$OS" in
    debian)  echo "sudo apt install make" ;;
    fedora)  echo "sudo dnf install make" ;;
    macos)   echo "xcode-select --install" ;;
    windows) echo "choco install make  OR  winget install ezwinports.make" ;;
    *)       echo "install GNU Make for your platform" ;;
  esac
}

docker_hint() {
  case "$OS" in
    debian|fedora|linux) echo "https://docs.docker.com/engine/install/" ;;
    macos|windows)       echo "https://docs.docker.com/desktop/" ;;
    *)                   echo "https://docs.docker.com/get-docker/" ;;
  esac
}

go_hint() { echo "https://go.dev/dl/"; }
node_hint() { echo "https://nodejs.org/"; }

# --- Check a command exists ---

check_required() {
  cmd="$1"
  hint="$2"
  if command -v "$cmd" >/dev/null 2>&1; then
    pass "$cmd"
  else
    fail "$cmd" "$hint"
    required_ok=false
  fi
}

check_dev() {
  cmd="$1"
  hint="$2"
  if command -v "$cmd" >/dev/null 2>&1; then
    pass "$cmd"
  else
    fail "$cmd" "$hint"
    dev_ok=false
  fi
}

# --- Required tools ---

printf "\n${BOLD}Checking required tools...${RESET}\n"
check_required "make" "$(make_hint)"
check_required "docker" "$(docker_hint)"

if docker compose version >/dev/null 2>&1; then
  pass "docker compose"
else
  fail "docker compose" "$(docker_hint)"
  required_ok=false
fi

# --- Development tools ---

printf "\n${BOLD}Checking development tools (needed for lint/test)...${RESET}\n"
check_dev "go" "$(go_hint)"
check_dev "node" "$(node_hint)"

# --- .env file ---

printf "\n${BOLD}Checking configuration...${RESET}\n"
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)

if [ -f "$PROJECT_DIR/.env" ]; then
  pass ".env file exists"
else
  if [ -f "$PROJECT_DIR/.env.example" ]; then
    cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    warn ".env created from .env.example" "edit it with your Discourse credentials"
  else
    fail ".env" ".env.example not found"
    required_ok=false
  fi
fi

# --- Summary ---

printf "\n"
if [ "$required_ok" = true ] && [ "$dev_ok" = true ]; then
  printf "${GREEN}All tools found.${RESET} Run ${BOLD}make start${RESET} to launch.\n"
elif [ "$required_ok" = true ]; then
  printf "${YELLOW}Required tools OK.${RESET} Some development tools are missing (see above).\n"
  printf "You can run ${BOLD}make start${RESET} now, but ${BOLD}make lint${RESET} and ${BOLD}make test${RESET} need the missing tools.\n"
else
  printf "${RED}Some required tools are missing.${RESET} Install them before running ${BOLD}make start${RESET}.\n"
  exit 1
fi
