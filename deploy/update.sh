#!/usr/bin/env bash
# Updates an existing medisite install (set up by deploy/install.sh) from
# this repo checkout: rebuilds the binary and replaces the installed one,
# and replaces templates/, static/, and content/ (posts + portfolio.yaml)
# under /opt/medisite with the repo's copies, then restarts the service.
#
# Unlike install.sh, this DOES overwrite content/ — the repo becomes the
# source of truth for posts and portfolio entries. Anything edited
# directly on the server is moved aside, not deleted: every replaced item
# (binary, templates/, static/, content/) is kept next to the live one
# with a .prev suffix, overwritten by the next update. To roll back:
#   sudo ./deploy/update.sh --rollback
#
# It does not touch the config, the systemd unit, or the service user —
# re-run install.sh for those.
#
# Usage: sudo ./deploy/update.sh [--rollback]

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR=/opt/medisite
SERVICE_USER=medisite
SERVICE_NAME=medisite.service
ITEMS=(medisite templates static content)

if [[ $EUID -ne 0 ]]; then
  echo "error: this script must be run as root (sudo $0)" >&2
  exit 1
fi

if [[ ! -x "$INSTALL_DIR/medisite" ]]; then
  echo "error: no existing install at $INSTALL_DIR — run deploy/install.sh first" >&2
  exit 1
fi

restart_and_check() {
  echo "==> Restarting $SERVICE_NAME"
  systemctl restart "$SERVICE_NAME"
  # medisite fails fast on bad content/templates, so give it a moment to
  # either come up or exit before checking.
  sleep 2
  if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "$SERVICE_NAME is active"
  else
    echo "error: $SERVICE_NAME is not active — check: journalctl -u $SERVICE_NAME -n 50" >&2
    return 1
  fi
}

if [[ "${1:-}" == "--rollback" ]]; then
  for item in "${ITEMS[@]}"; do
    if [[ ! -e "$INSTALL_DIR/$item.prev" ]]; then
      echo "error: $INSTALL_DIR/$item.prev not found, nothing to roll back to" >&2
      exit 1
    fi
  done
  echo "==> Rolling back to the previous version"
  for item in "${ITEMS[@]}"; do
    rm -rf "${INSTALL_DIR:?}/$item.failed"
    mv "$INSTALL_DIR/$item" "$INSTALL_DIR/$item.failed"
    mv "$INSTALL_DIR/$item.prev" "$INSTALL_DIR/$item"
    mv "$INSTALL_DIR/$item.failed" "$INSTALL_DIR/$item.prev"
  done
  restart_and_check
  exit $?
fi

# Stage everything next to the live files first, so a failed build or copy
# leaves the running install untouched, and the swap below is just renames.
STAGE_DIR="$(mktemp -d "$INSTALL_DIR/.update.XXXXXX")"
trap 'rm -rf "$STAGE_DIR"' EXIT

echo "==> Building medisite in $REPO_DIR"
( cd "$REPO_DIR" && go build -o "$STAGE_DIR/medisite" . )

echo "==> Staging templates, static assets, and content"
for dir in templates static content; do
  cp -r "$REPO_DIR/$dir" "$STAGE_DIR/$dir"
done
chown -R "$SERVICE_USER:$SERVICE_USER" "$STAGE_DIR"

# Renaming (rather than cp over) the binary also avoids "Text file busy"
# from overwriting an executable that is currently running.
echo "==> Swapping in the new version (previous kept as *.prev)"
for item in "${ITEMS[@]}"; do
  rm -rf "${INSTALL_DIR:?}/$item.prev"
  mv "$INSTALL_DIR/$item" "$INSTALL_DIR/$item.prev"
  mv "$STAGE_DIR/$item" "$INSTALL_DIR/$item"
done

if ! restart_and_check; then
  echo "hint: roll back with: sudo $0 --rollback" >&2
  exit 1
fi

echo "==> Done"
