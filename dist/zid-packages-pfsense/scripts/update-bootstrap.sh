#!/bin/sh
# zid-packages-update (bootstrap)

set -eu

URL_DEFAULT="https://s3.soulsolucoes.com.br/soul/portal/zid-packages-latest.tar.gz"
URL="${ZID_PACKAGES_UPDATE_URL:-$URL_DEFAULT}"
FORCE=0
KEEP_TMP=0

usage() {
	cat <<EOF2
ZID Packages updater (bootstrap)

Usage:
  sh /usr/local/sbin/zid-packages-update [-u <url>] [-f] [-k]

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

if [ "$(id -u)" != "0" ]; then
	echo "ERROR: This script must be run as root" >&2
	exit 1
fi

DOWNLOADER=""
if command -v fetch >/dev/null 2>&1; then
	DOWNLOADER="fetch"
elif command -v curl >/dev/null 2>&1; then
	DOWNLOADER="curl"
else
	echo "ERROR: Neither 'fetch' nor 'curl' found" >&2
	exit 1
fi

TMP_DIR="$(mktemp -d /tmp/zid-packages-update.XXXXXX)"
cleanup() {
	if [ "${KEEP_TMP}" -eq 1 ]; then
		echo "Keeping temp dir: ${TMP_DIR}"
		return
	fi
	rm -rf "${TMP_DIR}"
}
trap cleanup EXIT INT TERM

TARBALL="${TMP_DIR}/bundle.tar.gz"
EXTRACT_DIR="${TMP_DIR}/extract"
mkdir -p "${EXTRACT_DIR}"

echo "Downloading: ${URL}"
if [ "${DOWNLOADER}" = "fetch" ]; then
	fetch -o "${TARBALL}" "${URL}"
else
	curl -fL -o "${TARBALL}" "${URL}"
fi

echo "Extracting bundle..."
tar -xzf "${TARBALL}" -C "${EXTRACT_DIR}"

UPDATER_SH="$(find "${EXTRACT_DIR}" -maxdepth 5 -type f -path "*/scripts/update.sh" | head -n 1 || true)"
if [ -z "${UPDATER_SH}" ]; then
	echo "ERROR: update.sh not found inside bundle" >&2
	exit 1
fi

echo "Running bundled updater: ${UPDATER_SH}"
UPDATER_ARGS=""
if [ "${KEEP_TMP}" -eq 1 ]; then
	UPDATER_ARGS="${UPDATER_ARGS} -k"
fi
if [ "${URL}" != "${URL_DEFAULT}" ]; then
	UPDATER_ARGS="${UPDATER_ARGS} -u ${URL}"
fi
if [ "${FORCE}" -eq 1 ]; then
	UPDATER_ARGS="${UPDATER_ARGS} -f"
fi
sh "${UPDATER_SH}" ${UPDATER_ARGS}
