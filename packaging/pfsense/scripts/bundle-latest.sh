#!/bin/sh
set -e

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
VERSION_FILE="${ROOT_DIR}/zid-packages-latest.version"
TARBALL="${ROOT_DIR}/zid-packages-latest.tar.gz"
SHA_FILE="${ROOT_DIR}/sha256.txt"

VERSION="$(grep '^VERSION=' "${ROOT_DIR}/Makefile" | cut -d'=' -f2 | tr -d '[:space:]')"
if [ -z "${VERSION}" ]; then
  VERSION="dev"
fi

DIST_DIR="${ROOT_DIR}/dist/zid-packages-pfsense"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

if command -v go >/dev/null 2>&1; then
  mkdir -p "${ROOT_DIR}/build"
  (
    cd "${ROOT_DIR}"
    env GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 \
      go build -o "build/zid-packages" -ldflags "-X main.version=${VERSION}" ./cmd/zid-packages
  )
fi

cp -R "${ROOT_DIR}/packaging/pfsense/files" "${DIST_DIR}/"
cp -R "${ROOT_DIR}/packaging/pfsense/scripts" "${DIST_DIR}/"

if [ -f "${ROOT_DIR}/build/zid-packages" ]; then
  mkdir -p "${DIST_DIR}/build"
  cp "${ROOT_DIR}/build/zid-packages" "${DIST_DIR}/build/"
fi

printf "%s\n" "${VERSION}" > "${VERSION_FILE}"

tar -czf "${TARBALL}" -C "${ROOT_DIR}/dist" "zid-packages-pfsense"

if command -v sha256 >/dev/null 2>&1; then
  sha256 -q "${TARBALL}" > "${SHA_FILE}"
else
  sha256sum "${TARBALL}" | awk '{print $1}' > "${SHA_FILE}"
fi
