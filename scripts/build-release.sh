#!/usr/bin/env bash
set -euo pipefail

# Build release binaries for all supported architectures.
# Output: dist/release/vibecodepc-{arm64,arm,amd64}

BOLD="\033[1m"
DIM="\033[2m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="${ROOT_DIR}/dist/release"
VERSION=$(grep -m1 'const version' "${ROOT_DIR}/server/main.go" | sed 's/.*"\(.*\)".*/\1/')

echo ""
echo -e "  ${BOLD}Building VibeCodePC v${VERSION} release binaries${RESET}"
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""

# ── 1. Build client ────────────────────────────────────────────────────────
echo -e "  ${DIM}[1/3]${RESET}  Building client..."
cd "${ROOT_DIR}/client"
pnpm install --frozen-lockfile 2>/dev/null || pnpm install
pnpm build
echo -e "  ${GREEN}✓${RESET}  Client built"

# ── 2. Prepare release directory ───────────────────────────────────────────
mkdir -p "${RELEASE_DIR}"

# ── 3. Cross-compile for each target ──────────────────────────────────────
TARGETS=(
  "linux:arm64"
  "linux:arm"
  "linux:amd64"
)

LDFLAGS="-s -w -X main.version=${VERSION}"

echo -e "  ${DIM}[2/3]${RESET}  Compiling binaries..."

for target in "${TARGETS[@]}"; do
  GOOS="${target%%:*}"
  GOARCH="${target##*:}"

  # For arm (32-bit), set GOARM=7 for Raspberry Pi compatibility
  GOARM_ENV=""
  if [ "${GOARCH}" = "arm" ]; then
    GOARM_ENV="GOARM=7"
  fi

  BIN_NAME="vibecodepc-${GOARCH}"
  echo -ne "         ${DIM}${GOOS}/${GOARCH}...${RESET}"

  cd "${ROOT_DIR}"
  if env CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" ${GOARM_ENV} \
    go build -ldflags "${LDFLAGS}" -o "${RELEASE_DIR}/${BIN_NAME}" ./server/main.go; then
    SIZE=$(du -h "${RELEASE_DIR}/${BIN_NAME}" | cut -f1 | xargs)
    echo -e "  ${GREEN}✓${RESET}  ${SIZE}"
  else
    echo -e "  ${RED}✗${RESET}  failed"
    exit 1
  fi
done

# ── 4. Generate checksums ─────────────────────────────────────────────────
echo -ne "  ${DIM}[3/3]${RESET}  Generating checksums..."
cd "${RELEASE_DIR}"
if command -v sha256sum &>/dev/null; then
  sha256sum vibecodepc-* > checksums.txt
elif command -v shasum &>/dev/null; then
  shasum -a 256 vibecodepc-* > checksums.txt
fi
echo -e "  ${GREEN}✓${RESET}"

echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo -e "  ${GREEN}${BOLD}Release binaries ready${RESET}  →  ${DIM}${RELEASE_DIR}/${RESET}"
echo ""
ls -lh "${RELEASE_DIR}/"
echo ""
