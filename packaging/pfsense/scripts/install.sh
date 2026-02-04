#!/bin/sh
#
# ZID Packages pfSense Package Installation Script
#
# This script installs the ZID Packages package files to a pfSense system.
# Run this script on the pfSense firewall after copying the package files.
#
# Usage: ./install.sh
#

set -e

INSTALLER_VERSION="0.4.13"

echo "========================================="
echo " ZID Packages pfSense Package Installer"
echo "========================================="
echo "[INFO] Installer version: ${INSTALLER_VERSION}"

# Check if running as root
if [ "$(id -u)" != "0" ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Define paths
PREFIX="/usr/local"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
FILES_DIR="${ROOT_DIR}/files"

PHP_BIN="/usr/local/bin/php"
PHP_CMD="${PHP_BIN}"
if [ ! -x "${PHP_BIN}" ] || [ ! -s "${PHP_BIN}" ]; then
    if [ -x /usr/local/bin/php-cgi ]; then
        PHP_BIN="/usr/local/bin/php-cgi"
        PHP_CMD="${PHP_BIN} -q"
    else
        PHP_CMD=""
    fi
fi


echo ""
echo "Installing from: ${ROOT_DIR}"
echo ""

# Create directories
echo "Creating directories..."
mkdir -p ${PREFIX}/pkg
mkdir -p ${PREFIX}/www
mkdir -p ${PREFIX}/etc/rc.d
mkdir -p ${PREFIX}/sbin
mkdir -p ${PREFIX}/share/pfSense-pkg-zid-packages
mkdir -p /etc/inc/priv
mkdir -p /var/log
mkdir -p /var/db/zid-packages

set_rc_conf() {
	key="$1"
	val="$2"
	file="$3"
	touch "${file}"
	if grep -q "^${key}=" "${file}"; then
		sed -i '' "s|^${key}=.*|${key}=\"${val}\"|" "${file}"
	else
		echo "${key}=\"${val}\"" >> "${file}"
	fi
}

ensure_local_startup() {
	file="/etc/rc.conf.local"
	key="local_startup"
	dir="/usr/local/etc/rc.d"

	touch "${file}"
	if grep -q "^${key}=" "${file}"; then
		val=$(grep "^${key}=" "${file}" | tail -n 1 | sed 's/^local_startup=//; s/"//g')
		if echo "${val}" | grep -q "${dir}"; then
			return 0
		fi
		val=$(echo "${val} ${dir}" | tr -s ' ')
		sed -i '' "s|^${key}=.*|${key}=\"${val}\"|" "${file}"
	else
		echo "${key}=\"${dir}\"" >> "${file}"
	fi
}

# Install package configuration
echo "Installing package configuration..."
cp ${FILES_DIR}${PREFIX}/pkg/zid-packages.xml ${PREFIX}/pkg/
cp ${FILES_DIR}${PREFIX}/pkg/zid-packages.inc ${PREFIX}/pkg/

# Install web pages
echo "Installing web pages..."
cp -f ${FILES_DIR}${PREFIX}/www/zid-packages_packages.php ${PREFIX}/www/
cp -f ${FILES_DIR}${PREFIX}/www/zid-packages_services.php ${PREFIX}/www/
cp -f ${FILES_DIR}${PREFIX}/www/zid-packages_licensing.php ${PREFIX}/www/
cp -f ${FILES_DIR}${PREFIX}/www/zid-packages_logs.php ${PREFIX}/www/

# Install privilege definitions
echo "Installing privilege definitions..."
cp -f ${FILES_DIR}/etc/inc/priv/zid-packages.priv.inc /etc/inc/priv/

# Install rc.d script
echo "Installing rc.d..."
cp -f ${FILES_DIR}${PREFIX}/etc/rc.d/zid_packages ${PREFIX}/etc/rc.d/
chmod 755 ${PREFIX}/etc/rc.d/zid_packages
if [ -f ${FILES_DIR}${PREFIX}/etc/rc.d/zid_packages.sh ]; then
    cp -f ${FILES_DIR}${PREFIX}/etc/rc.d/zid_packages.sh ${PREFIX}/etc/rc.d/
    chmod 755 ${PREFIX}/etc/rc.d/zid_packages.sh
fi

# Install package binary
BINARY_PATH="${ROOT_DIR}/build/zid-packages"
if [ -f "${BINARY_PATH}" ]; then
    echo "Installing binary..."
    TMP_BIN="${PREFIX}/sbin/.zid-packages.new.$$"
    cp "${BINARY_PATH}" "${TMP_BIN}"
    chmod 755 "${TMP_BIN}"
    mv -f "${TMP_BIN}" "${PREFIX}/sbin/zid-packages"
    chmod 755 ${PREFIX}/sbin/zid-packages
else
    echo "Warning: Binary not found at ${BINARY_PATH}"
    echo "         You need to copy the zid-packages binary to ${PREFIX}/sbin/ manually"
fi

# Create log file
touch /var/log/zid-packages.log
chmod 644 /var/log/zid-packages.log

# Install updater helper (so future updates don't require manual tar/scp)
if [ -f "${SCRIPT_DIR}/update-bootstrap.sh" ]; then
    echo "Installing updater helper..."
    TMP_UPDATER="${PREFIX}/sbin/.zid-packages-update.new.$$"
    cp "${SCRIPT_DIR}/update-bootstrap.sh" "${TMP_UPDATER}"
    chmod 755 "${TMP_UPDATER}"
    mv -f "${TMP_UPDATER}" "${PREFIX}/sbin/zid-packages-update"

    TMP_UPDATER_INFO="${PREFIX}/share/pfSense-pkg-zid-packages/.zid-packages-update.new.$$"
    cp "${SCRIPT_DIR}/update-bootstrap.sh" "${TMP_UPDATER_INFO}"
    chmod 755 "${TMP_UPDATER_INFO}"
    mv -f "${TMP_UPDATER_INFO}" "${PREFIX}/share/pfSense-pkg-zid-packages/zid-packages-update"
fi

# Install register/unregister scripts
if [ -f "${SCRIPT_DIR}/register-package.php" ]; then
    cp "${SCRIPT_DIR}/register-package.php" "${PREFIX}/share/pfSense-pkg-zid-packages/register-package.php"
    chmod 755 "${PREFIX}/share/pfSense-pkg-zid-packages/register-package.php"
fi
if [ -f "${SCRIPT_DIR}/unregister-package.php" ]; then
    cp "${SCRIPT_DIR}/unregister-package.php" "${PREFIX}/share/pfSense-pkg-zid-packages/unregister-package.php"
    chmod 755 "${PREFIX}/share/pfSense-pkg-zid-packages/unregister-package.php"
fi

echo ""
echo "========================================="
echo " File Installation Complete!"
echo "========================================="
echo ""

# Register package
if [ -f "${SCRIPT_DIR}/register-package.php" ]; then
    echo "========================================="
    echo " Package Registration"
    echo "========================================="
    echo ""
    echo "Registering package with pfSense..."

    if [ -n "${PHP_CMD}" ]; then
        ${PHP_CMD} "${SCRIPT_DIR}/register-package.php"
        register_result=$?
    else
        echo "[ERROR] PHP nao encontrado (php/php-cgi)"
        register_result=1
    fi
    echo ""

    if [ $register_result -eq 0 ]; then
        echo "[OK] Package registered successfully"
        echo ""

        echo "Reloading pfSense web GUI (to pick up updated PHP pages)..."
        if [ -x /usr/local/sbin/pfSsh.php ]; then
            /usr/local/sbin/pfSsh.php playback reloadwebgui >/dev/null 2>&1 || true
        elif [ -x /etc/rc.restart_webgui ]; then
            /etc/rc.restart_webgui >/dev/null 2>&1 || true
        elif [ -x /usr/local/etc/rc.d/php-fpm ]; then
            /usr/local/etc/rc.d/php-fpm restart >/dev/null 2>&1 || true
        fi
        echo "[OK] Web GUI reload requested"
        echo ""
        echo "IMPORTANT: Wait ~10 seconds, then reload your browser (Ctrl+Shift+R)"
        echo ""
    else
        echo "[ERROR] Package registration failed!"
        echo "        You can try running manually:"
        echo "        php ${SCRIPT_DIR}/register-package.php"
        echo ""
    fi
else
    echo "[ERROR] register-package.php not found"
    echo ""
fi

# Configure rc/cron and start daemon
echo "Configuring rc/cron..."
if [ -n "${PHP_CMD}" ]; then
    TMP_PHP="/tmp/zid-packages-rc.$$.php"
    cat > "${TMP_PHP}" <<'PHP'
<?php
require_once("/usr/local/pkg/zid-packages.inc");
zid_packages_set_rc_enable(true);
zid_packages_remove_legacy_watchdogs();
zid_packages_install_watchdog_cron();
?>
PHP
    ${PHP_CMD} "${TMP_PHP}" >/dev/null 2>&1 || true
    rm -f "${TMP_PHP}"
else
	echo "[WARN] PHP nao encontrado para configurar rc/cron"
fi

# Ensure rc.conf.local has zid_packages_enable=YES for boot persistence
set_rc_conf "zid_packages_enable" "YES" "/etc/rc.conf.local"
# Ensure local package rc.d startup is enabled and includes /usr/local/etc/rc.d
set_rc_conf "localpkg_enable" "YES" "/etc/rc.conf.local"
ensure_local_startup

if [ -x /usr/local/etc/rc.d/zid_packages ]; then
    echo "Restarting zid-packages daemon..."
    /usr/local/etc/rc.d/zid_packages onerestart || true
fi

echo "========================================="
echo " Installation Summary"
echo "========================================="
echo ""
echo "Files installed:"
echo "  • Binary: ${PREFIX}/sbin/zid-packages"
echo "  • Package files: ${PREFIX}/pkg/zid-packages.*"
echo "  • Web interface: ${PREFIX}/www/zid-packages_*.php"
echo "  • RC script: ${PREFIX}/etc/rc.d/zid_packages"
if [ -f ${PREFIX}/etc/rc.d/zid_packages.sh ]; then
    echo "  • RC script (localpkg): ${PREFIX}/etc/rc.d/zid_packages.sh"
fi
echo "  • Updater: ${PREFIX}/sbin/zid-packages-update"
echo ""
echo "Next steps:"
echo ""
echo "1. Test the service:"
echo "   /usr/local/etc/rc.d/zid_packages status"
echo ""
echo "2. Access pfSense web interface:"
echo "   - Navigate to Services > ZID Packages"
echo ""
