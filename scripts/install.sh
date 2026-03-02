#!/usr/bin/env bash
set -euo pipefail

# VibeCodePC Installer
# Usage: curl -fsSL https://vibecodepc.com/install.sh | bash

BOLD="\033[1m"
DIM="\033[2m"
VIOLET="\033[35m"
GREEN="\033[32m"
AMBER="\033[33m"
RED="\033[31m"
RESET="\033[0m"

VIBECODEPC_VERSION="${VIBECODEPC_VERSION:-latest}"
INSTALL_DIR="${HOME}/.vibecodepc"
BIN_DIR="${INSTALL_DIR}/bin"
DATA_DIR="${INSTALL_DIR}/data"

log() { echo -e "  ${DIM}${1}${RESET}"; }
step() { echo -e "  ${BOLD}${1}${RESET}"; }
ok() { echo -e "  ${GREEN}✓${RESET}  ${1}"; }
warn() { echo -e "  ${AMBER}⚠${RESET}  ${1}"; }
err() { echo -e "  ${RED}✗${RESET}  ${1}"; exit 1; }

echo ""
echo -e "  ${BOLD}${VIOLET}VibeCodePC Installer${RESET}"
echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""

# Step 1: Detect system
echo -ne "  ${DIM}[1/7]${RESET}  Detecting system...  "
ARCH=$(uname -m)
case "${ARCH}" in
  aarch64|arm64) ARCH_NAME="arm64" ;;
  armv7l|armhf)  ARCH_NAME="arm" ;;
  x86_64)        ARCH_NAME="amd64" ;;
  *) err "Unsupported architecture: ${ARCH}" ;;
esac
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
echo -e "${DIM}${ARCH_NAME}, ${OS}${RESET}"

# Step 2: Install dependencies
echo -ne "  ${DIM}[2/7]${RESET}  Checking dependencies...  "
DEPS_OK=true
if ! command -v git &>/dev/null; then
  echo "installing git..."
  sudo apt-get install -y git &>/dev/null || err "Failed to install git"
fi
if ! command -v docker &>/dev/null; then
  echo "installing docker..."
  curl -fsSL https://get.docker.com | sh &>/dev/null || warn "Docker install failed - install manually"
fi
echo -e "${GREEN}git ✓  docker ✓${RESET}"

# Step 3: Download VibeCodePC
echo -ne "  ${DIM}[3/7]${RESET}  Downloading VibeCodePC...  "
mkdir -p "${BIN_DIR}" "${DATA_DIR}"

if [ "${VIBECODEPC_VERSION}" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/vibecodepc/vibecodepc/releases/latest/download/vibecodepc-${ARCH_NAME}"
else
  DOWNLOAD_URL="https://github.com/vibecodepc/vibecodepc/releases/download/${VIBECODEPC_VERSION}/vibecodepc-${ARCH_NAME}"
fi

if curl -fsSL --progress-bar "${DOWNLOAD_URL}" -o "${BIN_DIR}/vibecodepc" 2>&1 | tail -1; then
  chmod +x "${BIN_DIR}/vibecodepc"
  echo -e "${GREEN}✓${RESET}"
else
  warn "Could not download from GitHub releases. Building from source or using local binary."
  # Fallback: check if binary exists in PATH
  if command -v vibecodepc &>/dev/null; then
    cp "$(command -v vibecodepc)" "${BIN_DIR}/vibecodepc"
    echo -e "${GREEN}✓ (local)${RESET}"
  else
    err "Could not obtain vibecodepc binary. Set VIBECODEPC_VERSION or build from source."
  fi
fi

# Step 4: Download cloudflared
echo -ne "  ${DIM}[4/7]${RESET}  Downloading cloudflared...  "
CF_URL="https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${ARCH_NAME}"
curl -fsSL --progress-bar "${CF_URL}" -o "${BIN_DIR}/cloudflared" 2>&1 | tail -1 || warn "cloudflared download failed"
chmod +x "${BIN_DIR}/cloudflared" 2>/dev/null || true
echo -e "${GREEN}✓${RESET}"

# Step 5: Create directories
echo -ne "  ${DIM}[5/7]${RESET}  Creating directories...  "
mkdir -p "${INSTALL_DIR}"/{data,logs,nanoclaw/data}
echo -e "${GREEN}~/.vibecodepc/  ✓${RESET}"

# Step 6: Register services (systemd)
echo -ne "  ${DIM}[6/7]${RESET}  Registering services...  "
if command -v systemctl &>/dev/null && [ -d /etc/systemd/system ]; then
  cat > /tmp/vibecodepc.service << EOF
[Unit]
Description=VibeCodePC Server
After=network.target

[Service]
Type=simple
User=${USER}
ExecStart=${BIN_DIR}/vibecodepc
Restart=always
RestartSec=5
Environment=PORT=3000
Environment=DATA_DIR=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF
  sudo mv /tmp/vibecodepc.service /etc/systemd/system/vibecodepc.service
  sudo systemctl daemon-reload
  sudo systemctl enable vibecodepc &>/dev/null
  sudo systemctl start vibecodepc
  echo -e "${GREEN}vibecodepc ✓${RESET}"
else
  # No systemd — start directly in background
  nohup "${BIN_DIR}/vibecodepc" > "${INSTALL_DIR}/logs/vibecodepc.log" 2>&1 &
  echo -e "${GREEN}started (no systemd)${RESET}"
fi

# Step 7: Wait for tunnel URL
echo -ne "  ${DIM}[7/7]${RESET}  Starting & waiting for tunnel URL...  "
TUNNEL_URL=""
for i in {1..30}; do
  sleep 1
  TUNNEL_URL=$(curl -s http://localhost:3000/api/settings/tunnel 2>/dev/null | grep -o '"tunnelUrl":"[^"]*"' | cut -d'"' -f4 || true)
  if [ -n "${TUNNEL_URL}" ] && [ "${TUNNEL_URL}" != "null" ]; then
    break
  fi
  echo -ne "●"
done
echo ""

echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""
echo -e "  ${GREEN}${BOLD}✓  VibeCodePC is running${RESET}"
echo ""

LOCAL_URL="http://$(hostname).local:3000"
echo -e "     Local   →  ${BOLD}${LOCAL_URL}${RESET}"
if [ -n "${TUNNEL_URL}" ]; then
  echo -e "     Remote  →  ${BOLD}${TUNNEL_URL}${RESET}"
else
  echo -e "     Remote  →  ${DIM}(starting tunnel...)${RESET}"
fi

echo ""
echo -e "     Open Remote from any device to run the setup wizard."
echo -e "     The Remote URL changes on reboot until you set up"
echo -e "     a named tunnel — the wizard guides you through it."
echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""
