#!/usr/bin/env bash
# Replaces the site content (content/: posts + portfolio.yaml) of an
# existing medisite install (set up by deploy/install.sh) with this repo
# checkout's copy, then restarts the service so it's reloaded.
#
# Unlike install.sh, this DOES overwrite content/ — the repo becomes the
# source of truth for posts and portfolio entries. Anything edited
# directly on the server is moved aside, not deleted: the replaced
# content/ is kept next to the live one as content.prev, overwritten by
# the next update. To roll back:
#   sudo ./deploy/update.sh --rollback
#
# It touches nothing but content/ — the binary, templates/, static/,
# config, systemd unit, and service user are all install.sh's job
# (re-run it for code, template, CSS, or config changes).
#
# Usage: sudo ./deploy/update.sh [--rollback]

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR=/opt/medisite
SERVICE_USER=medisite
SERVICE_NAME=medisite.service
CONTENT="$INSTALL_DIR/content"

if [[ $EUID -ne 0 ]]; then
  echo "error: this script must be run as root (sudo $0)" >&2
  exit 1
fi

if [[ ! -d "$CONTENT" ]]; then
  echo "error: no existing install at $INSTALL_DIR — run deploy/install.sh first" >&2
  exit 1
fi

restart_and_check() {
  echo "==> Restarting $SERVICE_NAME"
  systemctl restart "$SERVICE_NAME"
  # medisite fails fast on bad content, so give it a moment to either come
  # up or exit before checking.
  sleep 2
  if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "$SERVICE_NAME is active"
  else
    echo "error: $SERVICE_NAME is not active — check: journalctl -u $SERVICE_NAME -n 50" >&2
    return 1
  fi
}

if [[ "${1:-}" == "--rollback" ]]; then
  if [[ ! -e "$CONTENT.prev" ]]; then
    echo "error: $CONTENT.prev not found, nothing to roll back to" >&2
    exit 1
  fi
  echo "==> Rolling back to the previous content"
  rm -rf "$CONTENT.failed"
  mv "$CONTENT" "$CONTENT.failed"
  mv "$CONTENT.prev" "$CONTENT"
  mv "$CONTENT.failed" "$CONTENT.prev"
  restart_and_check
  exit $?
fi

# Stage the new content next to the live copy first, so a failed copy
# leaves the running install untouched, and the swap below is just renames.
STAGE_DIR="$(mktemp -d "$INSTALL_DIR/.update.XXXXXX")"
trap 'rm -rf "$STAGE_DIR"' EXIT

echo "==> Staging content"
cp -r "$REPO_DIR/content" "$STAGE_DIR/content"
chown -R "$SERVICE_USER:$SERVICE_USER" "$STAGE_DIR"

echo "==> Swapping in the new content (previous kept as content.prev)"
rm -rf "$CONTENT.prev"
mv "$CONTENT" "$CONTENT.prev"
mv "$STAGE_DIR/content" "$CONTENT"

if ! restart_and_check; then
  echo "hint: roll back with: sudo $0 --rollback" >&2
  exit 1
fi

echo "==> Done"
