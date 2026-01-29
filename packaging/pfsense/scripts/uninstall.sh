#!/bin/sh
# ZID Packages pfSense Package Uninstall

set -e

if [ "$(id -u)" != "0" ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

PREFIX="/usr/local"
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

if [ -x /usr/local/etc/rc.d/zid_packages ]; then
    /usr/local/etc/rc.d/zid_packages stop || true
fi

if [ -f "${PREFIX}/share/pfSense-pkg-zid-packages/unregister-package.php" ]; then
    if [ -n "${PHP_CMD}" ]; then
        ${PHP_CMD} "${PREFIX}/share/pfSense-pkg-zid-packages/unregister-package.php" || true
    fi
fi

if [ -n "${PHP_CMD}" ]; then
    TMP_PHP="/tmp/zid-packages-uninstall.$$.php"
    cat > "${TMP_PHP}" <<'PHP'
<?php
require_once("/usr/local/pkg/zid-packages.inc");
zid_packages_remove_watchdog_cron();
zid_packages_remove_legacy_watchdogs();
?>
PHP
    ${PHP_CMD} "${TMP_PHP}" >/dev/null 2>&1 || true
    rm -f "${TMP_PHP}"
fi

rm -f ${PREFIX}/pkg/zid-packages.xml
rm -f ${PREFIX}/pkg/zid-packages.inc
rm -f ${PREFIX}/www/zid-packages_packages.php
rm -f ${PREFIX}/www/zid-packages_services.php
rm -f ${PREFIX}/www/zid-packages_licensing.php
rm -f ${PREFIX}/www/zid-packages_logs.php
rm -f ${PREFIX}/www/services/zid-packages_packages.php
rm -f ${PREFIX}/www/services/zid-packages_services.php
rm -f ${PREFIX}/www/services/zid-packages_licensing.php
rm -f ${PREFIX}/www/services/zid-packages_logs.php
rm -f /etc/inc/priv/zid-packages.priv.inc
rm -f ${PREFIX}/etc/rc.d/zid_packages
rm -f ${PREFIX}/sbin/zid-packages
rm -f ${PREFIX}/sbin/zid-packages-update
rm -f ${PREFIX}/share/pfSense-pkg-zid-packages/zid-packages-update
rm -f ${PREFIX}/share/pfSense-pkg-zid-packages/register-package.php
rm -f ${PREFIX}/share/pfSense-pkg-zid-packages/unregister-package.php

rm -f /var/run/zid-packages.pid

if [ -x /etc/rc.restart_webgui ]; then
    /etc/rc.restart_webgui
fi

echo "Uninstall concluido."
