#!/usr/local/bin/php
<?php
/*
 * unregister-package.php
 */

echo "=========================================\n";
echo " ZID Packages Package Unregister\n";
echo "=========================================\n\n";

if (posix_geteuid() !== 0) {
    echo "Error: This script must be run as root\n";
    exit(1);
}

if (!file_exists('/etc/inc/config.inc')) {
    echo "Error: This does not appear to be a pfSense system\n";
    exit(1);
}

require_once('/etc/inc/config.inc');
require_once('/etc/inc/util.inc');

$config = parse_config(true);

if (!is_array($config['installedpackages'])) {
    echo "No installedpackages section\n";
    exit(0);
}

echo "Removing package entries (if any)...\n";
if (is_array($config['installedpackages']['package'])) {
    foreach ($config['installedpackages']['package'] as $idx => $pkg) {
        if (isset($pkg['name']) && $pkg['name'] === 'zid-packages') {
            unset($config['installedpackages']['package'][$idx]);
            echo "  - Removed package entry\n";
        }
    }
    $config['installedpackages']['package'] = array_values($config['installedpackages']['package']);
}

echo "Removing menu entries (if any)...\n";
if (is_array($config['installedpackages']['menu'])) {
    foreach ($config['installedpackages']['menu'] as $idx => $menu) {
        if (isset($menu['name']) && $menu['name'] === 'ZID Packages') {
            unset($config['installedpackages']['menu'][$idx]);
            echo "  - Removed menu entry\n";
        }
    }
    $config['installedpackages']['menu'] = array_values($config['installedpackages']['menu']);
}

if (isset($config['installedpackages']['zidpackages'])) {
    unset($config['installedpackages']['zidpackages']);
}

write_config('ZID Packages package unregistered');

echo "  ✓ Configuration saved\n";

echo "\n=========================================\n";
echo " Unregister Complete!\n";
echo "=========================================\n\n";
?>
