#!/usr/bin/env bash
# Builds medisite and installs/updates it as a systemd service: the
# binary, templates/, and static/ under /opt/medisite, config.yaml under
# /etc/medisite, and logs under /var/log/medisite. Safe to re-run for
# redeploys (e.g. after a code, config, or static/template change) — it
# rebuilds the binary and restarts the service. It does NOT touch
# content/ once it exists, so it never clobbers posts or portfolio
# entries edited directly on the server; content/ is only seeded from
# the repo on the very first install.
#
# Usage: sudo ./deploy/install.sh

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR=/opt/medisite
CONFIG_DIR=/etc/medisite
LOG_DIR=/var/log/medisite
SERVICE_USER=medisite
SERVICE_NAME=medisite.service

if [[ $EUID -ne 0 ]]; then
  echo "error: this script must be run as root (sudo $0)" >&2
  exit 1
fi

echo "==> Building medisite in $REPO_DIR"
( cd "$REPO_DIR" && go build -o medisite . )

echo "==> Ensuring system user '$SERVICE_USER' exists"
if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
  useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

echo "==> Ensuring $INSTALL_DIR, $CONFIG_DIR, and $LOG_DIR exist"
mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"

echo "==> Installing binary"
cp "$REPO_DIR/medisite" "$INSTALL_DIR/medisite"

echo "==> Installing config"
cp "$REPO_DIR/config.yaml" "$CONFIG_DIR/config.yaml"

echo "==> Syncing templates and static assets"
for dir in templates static; do
  rm -rf "${INSTALL_DIR:?}/$dir"
  cp -r "$REPO_DIR/$dir" "$INSTALL_DIR/$dir"
done

if [[ -d "$INSTALL_DIR/content" ]]; then
  echo "==> $INSTALL_DIR/content already exists, leaving it alone (posts/portfolio are edited on the server, not redeployed)"
else
  echo "==> Seeding $INSTALL_DIR/content from the repo (first install)"
  cp -r "$REPO_DIR/content" "$INSTALL_DIR/content"
fi

echo "==> Setting ownership"
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR"

echo "==> Installing systemd service"
cp "$REPO_DIR/deploy/medisite.service" "/etc/systemd/system/$SERVICE_NAME"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

echo "==> Done"
if systemctl is-active --quiet "$SERVICE_NAME"; then
  echo "$SERVICE_NAME is active"
else
  echo "warning: $SERVICE_NAME is not active — check: journalctl -u $SERVICE_NAME -n 50" >&2
  exit 1
fi
