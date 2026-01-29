#!/bin/sh
# update.sh

set -e

URL_DEFAULT="https://s3.soulsolucoes.com.br/soul/portal/zid-packages-latest.tar.gz"
URL="${ZID_PACKAGES_UPDATE_URL:-$URL_DEFAULT}"
FORCE=0
KEEP_TMP=0

usage() {
	cat <<EOF2
ZID Packages updater

Usage:
  sh update.sh [-u <url>] [-f] [-k]

Options:
  -u <url>  Bundle URL (default: ${URL_DEFAULT})
  -f        Force update (skip version check)
  -k        Keep temporary directory (debug)
EOF2
}

while getopts "u:fkh" opt; do
	case "$opt" in
		u) URL="$OPTARG" ;;
		f) FORCE=1 ;;
		k) KEEP_TMP=1 ;;
		h) usage; exit 0 ;;
		*) usage; exit 2 ;;
	esac
done

PKG_DIR="$(dirname "$0")"

INSTALL_SH="$(find "${PKG_DIR}" -maxdepth 2 -type f -path "*/install.sh" | head -n 1 || true)"
if [ -z "${INSTALL_SH}" ]; then
	echo "ERROR: install.sh not found" >&2
	exit 1
fi

sh "${INSTALL_SH}"
