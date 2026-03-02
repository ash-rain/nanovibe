#!/usr/bin/env bash
set -euo pipefail

# VibeCodePC Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/ash-rain/nanovibe/main/scripts/install.sh | bash

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
LOG_DIR="${INSTALL_DIR}/logs"
GITHUB_REPO="vibecodepc/vibecodepc"

log()  { echo -e "     ${DIM}${1}${RESET}"; }
step() { echo -e "  ${BOLD}${1}${RESET}"; }
ok()   { echo -e "  ${GREEN}✓${RESET}  ${1}"; }
warn() { echo -e "  ${AMBER}⚠${RESET}  ${1}"; }
err()  { echo -e "  ${RED}✗${RESET}  ${1}"; exit 1; }

echo ""
echo -e "  ${BOLD}${VIOLET}VibeCodePC${RESET}  ${DIM}— AI coding station for Raspberry Pi${RESET}"
echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""

# ── 1. Detect system ────────────────────────────────────────────────────────
echo -ne "  ${DIM}[1/7]${RESET}  Detecting system..."
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

case "${ARCH}" in
  aarch64|arm64) BIN_ARCH="arm64"; CF_ARCH="arm64" ;;
  armv7l|armhf)  BIN_ARCH="arm";   CF_ARCH="arm"   ;;
  x86_64)        BIN_ARCH="amd64"; CF_ARCH="amd64"  ;;
  *) err "Unsupported architecture: ${ARCH}" ;;
esac

echo -e "  ${DIM}${ARCH} (${OS})${RESET}"

# ── 2. Install system dependencies ─────────────────────────────────────────
echo -ne "  ${DIM}[2/7]${RESET}  Checking dependencies..."

MISSING_DEPS=()
command -v curl &>/dev/null || MISSING_DEPS+=("curl")
command -v git  &>/dev/null || MISSING_DEPS+=("git")

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
  echo ""
  log "Installing: ${MISSING_DEPS[*]}"
  if command -v apt-get &>/dev/null; then
    sudo apt-get install -y "${MISSING_DEPS[@]}" -qq || err "Failed to install dependencies"
  elif command -v apk &>/dev/null; then
    sudo apk add "${MISSING_DEPS[@]}" >/dev/null || err "Failed to install dependencies"
  else
    err "Please install: ${MISSING_DEPS[*]}"
  fi
fi

if ! command -v docker &>/dev/null; then
  echo ""
  log "Installing Docker..."
  curl -fsSL https://get.docker.com | sh >/dev/null 2>&1 || warn "Docker install failed — install manually: https://docs.docker.com/engine/install/"
  if command -v docker &>/dev/null && ! groups | grep -q docker; then
    sudo usermod -aG docker "${USER}" || true
    warn "Added ${USER} to docker group — re-login may be needed for Docker without sudo"
  fi
fi

echo -e "  ${GREEN}ok${RESET}"

# ── 3. Download VibeCodePC binary ───────────────────────────────────────────
echo -ne "  ${DIM}[3/7]${RESET}  Downloading VibeCodePC..."
mkdir -p "${BIN_DIR}" "${DATA_DIR}" "${LOG_DIR}"

if [ "${VIBECODEPC_VERSION}" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/vibecodepc-${BIN_ARCH}"
else
  DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VIBECODEPC_VERSION}/vibecodepc-${BIN_ARCH}"
fi

if curl -fsSL "${DOWNLOAD_URL}" -o "${BIN_DIR}/vibecodepc" 2>/dev/null; then
  chmod +x "${BIN_DIR}/vibecodepc"
  echo -e "  ${GREEN}ok${RESET}"
else
  # Check if a local binary is already present (re-install case)
  if [ -x "${BIN_DIR}/vibecodepc" ]; then
    echo -e "  ${DIM}(kept existing)${RESET}"
  elif command -v vibecodepc &>/dev/null; then
    cp "$(command -v vibecodepc)" "${BIN_DIR}/vibecodepc"
    echo -e "  ${DIM}(from PATH)${RESET}"
  else
    err "Could not download vibecodepc. Check your internet connection or set VIBECODEPC_VERSION."
  fi
fi

# ── 4. Download cloudflared ─────────────────────────────────────────────────
echo -ne "  ${DIM}[4/7]${RESET}  Downloading cloudflared..."
CF_URL="https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CF_ARCH}"

if curl -fsSL "${CF_URL}" -o "${BIN_DIR}/cloudflared" 2>/dev/null; then
  chmod +x "${BIN_DIR}/cloudflared"
  echo -e "  ${GREEN}ok${RESET}"
else
  warn "Could not download cloudflared — tunnel will be unavailable"
fi

# Symlink cloudflared to /usr/local/bin so child processes can find it
if [ -x "${BIN_DIR}/cloudflared" ]; then
  sudo ln -sf "${BIN_DIR}/cloudflared" /usr/local/bin/cloudflared 2>/dev/null || true
fi

# ── 5. Install helper scripts ────────────────────────────────────────────────
echo -ne "  ${DIM}[5/7]${RESET}  Installing helper scripts..."

# update.sh — called by the in-app "Update" button
cat > "${INSTALL_DIR}/update.sh" << 'UPDATESCRIPT'
#!/usr/bin/env bash
set -euo pipefail
INSTALL_DIR="${HOME}/.vibecodepc"
BIN_DIR="${INSTALL_DIR}/bin"

echo "Checking for updates..."
ARCH=$(uname -m)
case "${ARCH}" in
  aarch64|arm64) BIN_ARCH="arm64" ;;
  armv7l|armhf)  BIN_ARCH="arm"   ;;
  x86_64)        BIN_ARCH="amd64" ;;
  *) echo "Unsupported arch: ${ARCH}"; exit 1 ;;
esac

DOWNLOAD_URL="https://github.com/vibecodepc/vibecodepc/releases/latest/download/vibecodepc-${BIN_ARCH}"
echo "Downloading new binary..."
curl -fsSL "${DOWNLOAD_URL}" -o "${BIN_DIR}/vibecodepc.new"
chmod +x "${BIN_DIR}/vibecodepc.new"
mv "${BIN_DIR}/vibecodepc.new" "${BIN_DIR}/vibecodepc"
echo "Restarting service..."
systemctl --user restart vibecodepc 2>/dev/null || sudo systemctl restart vibecodepc 2>/dev/null || true
echo "done: update complete"
UPDATESCRIPT

chmod +x "${INSTALL_DIR}/update.sh"

# uninstall.sh
cat > "${INSTALL_DIR}/uninstall.sh" << 'UNINSTALLSCRIPT'
#!/usr/bin/env bash
set -euo pipefail

echo ""
echo "This will remove VibeCodePC and its config."
echo "Your project directories will NOT be touched."
echo ""
read -p "Continue? [y/N] " confirm
if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

echo "Stopping services..."
sudo systemctl stop vibecodepc 2>/dev/null || pkill vibecodepc 2>/dev/null || true
sudo systemctl disable vibecodepc 2>/dev/null || true
sudo rm -f /etc/systemd/system/vibecodepc.service
sudo rm -f /usr/local/bin/cloudflared
sudo systemctl daemon-reload 2>/dev/null || true

echo "Removing binaries and logs..."
rm -rf "${HOME}/.vibecodepc/bin"
rm -rf "${HOME}/.vibecodepc/logs"
# data/ is preserved — user data is safe

echo ""
echo "VibeCodePC removed. Your projects are safe."
echo ""
UNINSTALLSCRIPT

chmod +x "${INSTALL_DIR}/uninstall.sh"
echo -e "  ${GREEN}ok${RESET}"

# ── 6. Register as a system service ─────────────────────────────────────────
echo -ne "  ${DIM}[6/7]${RESET}  Registering service..."

SYS_PATH="${BIN_DIR}:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

if command -v systemctl &>/dev/null && [ -d /etc/systemd/system ]; then
  sudo tee /etc/systemd/system/vibecodepc.service > /dev/null << EOF
[Unit]
Description=VibeCodePC Server
Documentation=https://vibecodepc.com
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${USER}
ExecStart=${BIN_DIR}/vibecodepc
Restart=always
RestartSec=5
StandardOutput=append:${LOG_DIR}/vibecodepc.log
StandardError=append:${LOG_DIR}/vibecodepc.log
Environment=PORT=3000
Environment=HOST=0.0.0.0
Environment=DATA_DIR=${DATA_DIR}
Environment=PATH=${SYS_PATH}

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl daemon-reload
  sudo systemctl enable vibecodepc >/dev/null 2>&1
  sudo systemctl restart vibecodepc
  echo -e "  ${GREEN}ok (systemd)${RESET}"

else
  # No systemd — start in background with nohup
  pkill -f "${BIN_DIR}/vibecodepc" 2>/dev/null || true
  sleep 1
  DATA_DIR="${DATA_DIR}" PORT=3000 HOST=0.0.0.0 PATH="${SYS_PATH}" \
    nohup "${BIN_DIR}/vibecodepc" > "${LOG_DIR}/vibecodepc.log" 2>&1 &
  echo -e "  ${GREEN}ok (background)${RESET}"
fi

# ── 7. Wait for Cloudflare tunnel URL ───────────────────────────────────────
echo -ne "  ${DIM}[7/7]${RESET}  Waiting for tunnel URL"

TUNNEL_URL=""
MAX_WAIT=45
for i in $(seq 1 ${MAX_WAIT}); do
  sleep 2
  # Try to get the tunnel URL from the API
  RESPONSE=$(curl -s --max-time 2 http://localhost:3000/api/settings/tunnel 2>/dev/null || true)
  TUNNEL_URL=$(echo "${RESPONSE}" | grep -o '"tunnelUrl":"[^"]*"' | sed 's/"tunnelUrl":"//;s/"//' || true)
  if [ -n "${TUNNEL_URL}" ] && [ "${TUNNEL_URL}" != "null" ] && [ "${TUNNEL_URL}" != "" ]; then
    break
  fi
  echo -ne "."
done
echo ""

# ── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""
echo -e "  ${GREEN}${BOLD}VibeCodePC is running${RESET}"
echo ""

HOSTNAME_LOCAL=$(hostname 2>/dev/null || echo "raspberrypi")
LOCAL_URL="http://${HOSTNAME_LOCAL}.local:3000"

echo -e "     Local    →  ${BOLD}${LOCAL_URL}${RESET}"

if [ -n "${TUNNEL_URL}" ]; then
  echo -e "     Remote   →  ${BOLD}${TUNNEL_URL}${RESET}"
else
  echo -e "     Remote   →  ${DIM}(tunnel starting — check the wizard in ~30s)${RESET}"
fi

echo ""
echo -e "  Open the URL above to run the setup wizard."

if [ -z "${TUNNEL_URL}" ]; then
  echo -e "  The Remote URL will appear in the wizard once the tunnel is ready."
fi

echo ""
echo -e "  ${DIM}The Remote URL changes on reboot unless you configure a named"
echo -e "  tunnel in the wizard (Cloudflare Zero Trust — free tier).${RESET}"
echo ""
echo -e "  ${DIM}Logs: ${LOG_DIR}/vibecodepc.log${RESET}"
echo -e "  ${DIM}┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄${RESET}"
echo ""
