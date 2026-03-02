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
sudo systemctl daemon-reload 2>/dev/null || true

echo "Removing binary and app data..."
rm -rf "${HOME}/.vibecodepc/bin"
rm -rf "${HOME}/.vibecodepc/logs"
# Keep data/ directory — user data preserved

echo ""
echo "VibeCodePC removed. Your projects are safe — the app never touched them."
echo ""
