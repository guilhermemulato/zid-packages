#!/usr/local/bin/php
<?php
/*
 * register-package.php
 *
 * Registers the ZID Packages package in pfSense's config.xml.
 * This makes the package visible in the web interface AND enables auto-start on boot.
 *
 * Usage: php register-package.php
 */

echo "=========================================\n";
echo " ZID Packages Package Registration v1.0.0\n";
echo "=========================================\n\n";

// Check if running as root
if (posix_geteuid() !== 0) {
    echo "Error: This script must be run as root\n";
    exit(1);
}

// Check if this is actually pfSense
if (!file_exists('/etc/inc/config.inc')) {
    echo "Error: This does not appear to be a pfSense system\n";
    exit(1);
}

echo "Loading pfSense configuration system...\n";
require_once('/etc/inc/config.inc');
require_once('/etc/inc/util.inc');

function zidpackages_detect_version() {
    $bin = '/usr/local/sbin/zid-packages';
    if (!is_executable($bin)) {
        return 'unknown';
    }
    $out = array();
    $rc = 0;
    exec(escapeshellcmd($bin) . ' -version 2>&1', $out, $rc);
    if ($rc !== 0 || empty($out)) {
        return 'unknown';
    }
    foreach ($out as $line) {
        $line = trim($line);
        if ($line === '') {
            continue;
        }
        if (preg_match('/\b(\d+\.\d+(?:\.\d+){0,3})\b/', $line, $matches)) {
            return $matches[1];
        }
        return $line;
    }
    return 'unknown';
}

// Parse current configuration
echo "Parsing configuration...\n";
$config = parse_config(true);

// Initialize arrays if they don't exist
if (!is_array($config['installedpackages'])) {
    $config['installedpackages'] = array();
}
if (!is_array($config['installedpackages']['package'])) {
    $config['installedpackages']['package'] = array();
}
if (!is_array($config['installedpackages']['menu'])) {
    $config['installedpackages']['menu'] = array();
}
if (!is_array($config['installedpackages']['zidpackages'])) {
    $config['installedpackages']['zidpackages'] = array();
}
if (!is_array($config['installedpackages']['zidpackages']['config'])) {
    $config['installedpackages']['zidpackages']['config'] = array();
}

// Remove old package entries to avoid duplicates
echo "Removing old package entries (if any)...\n";
foreach ($config['installedpackages']['package'] as $idx => $pkg) {
    if (!isset($pkg['name'])) {
        continue;
    }
    if ($pkg['name'] === 'zid-packages' || $pkg['name'] === 'zidpackages') {
        unset($config['installedpackages']['package'][$idx]);
        echo "  - Removed old package entry\n";
    }
}
// Reindex array to avoid gaps
$config['installedpackages']['package'] = array_values($config['installedpackages']['package']);

$detected_version = zidpackages_detect_version();
if ($detected_version === 'unknown') {
    $detected_version = '0.4.4';
}

// Add package entry with correct tag names
echo "Adding package entry...\n";
$config['installedpackages']['package'][] = array(
    'name' => 'zid-packages',
    'version' => $detected_version,
    'descr' => 'ZID Packages - package manager',
    'website' => '',
    'configurationfile' => 'zid-packages.xml',
    'include_file' => '/usr/local/pkg/zid-packages.inc'
);
echo "  ✓ Package entry added\n";

// Remove old menu entries to avoid duplicates
echo "Removing old menu entries (if any)...\n";
foreach ($config['installedpackages']['menu'] as $idx => $menu) {
    if (isset($menu['name']) && $menu['name'] === 'ZID Packages') {
        unset($config['installedpackages']['menu'][$idx]);
        echo "  - Removed old menu entry\n";
    }
}
// Reindex array to avoid gaps
$config['installedpackages']['menu'] = array_values($config['installedpackages']['menu']);

// Add menu entry - THIS IS CRITICAL FOR BOTH MENU AND AUTO-START!
echo "Adding menu entry to config.xml...\n";
$config['installedpackages']['menu'][] = array(
    'name' => 'ZID Packages',
    'tooltiptext' => 'ZID Packages manager',
    'section' => 'Services',
    'url' => '/zid-packages_packages.php'
);
echo "  ✓ Menu entry added (this enables menu display AND boot auto-start)\n";

// Initialize default configuration if empty
if (empty($config['installedpackages']['zidpackages']['config'])) {
    echo "Creating default configuration...\n";
    $config['installedpackages']['zidpackages']['config'][0] = array(
        'enable' => 'on'
    );
    echo "  ✓ Default config created\n";
}

// Write configuration
echo "Writing configuration to /cf/conf/config.xml...\n";
write_config("ZID Packages package registered");
echo "  ✓ Configuration saved\n";

echo "\n=========================================\n";
echo " Registration Complete!\n";
echo "=========================================\n\n";

echo "✓ Package entry added to config.xml\n";
echo "✓ Menu entry added to config.xml\n";
echo "✓ Default configuration created\n\n";

echo "IMPORTANT: To see the menu in the web interface, reload the web GUI:\n\n";
echo "  /etc/rc.restart_webgui\n\n";
echo "  (Wait ~10 seconds for the GUI to reload)\n\n";

echo "Then reload your browser (Ctrl+Shift+R) and check Services > ZID Packages\n\n";

exit(0);
?>
