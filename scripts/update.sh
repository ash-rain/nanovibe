#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/.vibecodepc"
BIN_DIR="${INSTALL_DIR}/bin"
ARCH=$(uname -m)
case "${ARCH}" in
  aarch64|arm64) ARCH_NAME="arm64" ;;
  armv7l|armhf)  ARCH_NAME="arm" ;;
  x86_64)        ARCH_NAME="amd64" ;;
esac

echo "Downloading latest VibeCodePC..."
curl -fsSL "https://github.com/vibecodepc/vibecodepc/releases/latest/download/vibecodepc-${ARCH_NAME}" -o /tmp/vibecodepc-new
chmod +x /tmp/vibecodepc-new

echo "Stopping vibecodepc service..."
sudo systemctl stop vibecodepc 2>/dev/null || pkill vibecodepc 2>/dev/null || true

echo "Replacing binary..."
mv /tmp/vibecodepc-new "${BIN_DIR}/vibecodepc"

echo "Starting vibecodepc service..."
sudo systemctl start vibecodepc 2>/dev/null || nohup "${BIN_DIR}/vibecodepc" > "${INSTALL_DIR}/logs/vibecodepc.log" 2>&1 &

echo "Waiting for service to start..."
sleep 3
journalctl -u vibecodepc -n 20 --no-pager 2>/dev/null || tail -20 "${INSTALL_DIR}/logs/vibecodepc.log" 2>/dev/null || true

VERSION=$("${BIN_DIR}/vibecodepc" --version 2>/dev/null || echo "unknown")
echo "Updated to ${VERSION} ✓"
