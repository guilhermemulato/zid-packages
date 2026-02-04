<?php
/*
 * zid-packages.inc
 *
 * Helper functions for ZID Packages pfSense package.
 */

require_once("config.inc");
require_once("util.inc");

if (!function_exists('config_lock')) {
	function config_lock() {
		if (function_exists('lock')) {
			lock('config');
		}
	}
}
if (!function_exists('config_unlock')) {
	function config_unlock() {
		if (function_exists('unlock')) {
			unlock('config');
		}
	}
}

const ZID_PACKAGES_BIN = '/usr/local/sbin/zid-packages';
const ZID_PACKAGES_LOG = '/var/log/zid-packages.log';
const ZID_PACKAGES_WATCHDOG_CMD = '/usr/local/sbin/zid-packages watchdog --once';
const ZID_PACKAGES_WATCHDOG_DESCR = 'ZID Packages: watchdog';

function zid_packages_status_json(&$error = '') {
	if (!is_executable(ZID_PACKAGES_BIN)) {
		$error = 'zid-packages not installed.';
		return array();
	}
	$out = array();
	$rc = 0;
	exec(escapeshellcmd(ZID_PACKAGES_BIN) . ' status --json 2>/dev/null', $out, $rc);
	if ($rc !== 0 || empty($out)) {
		$error = 'Failed to run status.';
		return array();
	}
	$raw = implode("\n", $out);
	$data = json_decode($raw, true);
	if (!is_array($data)) {
		$error = 'Invalid JSON.';
		return array();
	}
	return $data;
}

function zid_packages_run_cmd($cmd, &$output = '', &$rc = 0) {
	$out = array();
	$rc = 0;
	exec($cmd, $out, $rc);
	$output = trim(implode("\n", $out));
	return $rc === 0;
}

function zid_packages_set_rc_enable($enabled) {
	$path = '/etc/rc.conf.local';
	$value = $enabled ? 'YES' : 'NO';
	$line = 'zid_packages_enable="' . $value . '"';
	$lines = array();
	$found = false;
	if (file_exists($path)) {
		$lines = file($path, FILE_IGNORE_NEW_LINES);
		if (!is_array($lines)) {
			$lines = array();
		}
	}
	foreach ($lines as $idx => $row) {
		if (preg_match('/^zid_packages_enable=/', $row)) {
			$lines[$idx] = $line;
			$found = true;
		}
	}
	if (!$found) {
		$lines[] = $line;
	}
	file_put_contents($path, implode("\n", $lines) . "\n");
	@chmod($path, 0644);
}

function zid_packages_configure_cron_compat() {
	if (function_exists('configure_cron')) {
		@configure_cron();
	}
}

function zid_packages_remove_cron_jobs_where($predicate) {
	global $config;
	$changed = false;
	foreach (array('item', 'job') as $key) {
		if (!isset($config['cron'][$key]) || !is_array($config['cron'][$key])) {
			continue;
		}
		$new = array();
		foreach ($config['cron'][$key] as $job) {
			if ($predicate($job)) {
				$changed = true;
				continue;
			}
			$new[] = $job;
		}
		$config['cron'][$key] = $new;
	}
	return $changed;
}

function zid_packages_upsert_cron_job($descr, $minute, $hour, $mday, $month, $wday, $who, $command) {
	global $config;
	$targets = array('item', 'job');
	foreach ($targets as $key) {
		if (isset($config['cron'][$key]) && is_array($config['cron'][$key])) {
			foreach ($config['cron'][$key] as $idx => $job) {
				$job_descr = (string)($job['descr'] ?? '');
				$job_cmd = (string)($job['command'] ?? '');
				if ($job_descr === $descr || $job_cmd === $command) {
					$config['cron'][$key][$idx]['minute'] = $minute;
					$config['cron'][$key][$idx]['hour'] = $hour;
					$config['cron'][$key][$idx]['mday'] = $mday;
					$config['cron'][$key][$idx]['month'] = $month;
					$config['cron'][$key][$idx]['wday'] = $wday;
					$config['cron'][$key][$idx]['who'] = $who;
					$config['cron'][$key][$idx]['command'] = $command;
					$config['cron'][$key][$idx]['descr'] = $descr;
					$config['cron'][$key][$idx]['enabled'] = 'true';
					return true;
				}
			}
		}
	}
	$config['cron'] = is_array($config['cron'] ?? null) ? $config['cron'] : [];
	if (!isset($config['cron']['item']) || !is_array($config['cron']['item'])) {
		$config['cron']['item'] = array();
	}
	$config['cron']['item'][] = [
		'minute' => $minute,
		'hour' => $hour,
		'mday' => $mday,
		'month' => $month,
		'wday' => $wday,
		'who' => $who,
		'command' => $command,
		'descr' => $descr,
		'enabled' => 'true',
	];
	return true;
}

function zid_packages_watchdog_cron_command() {
	return sprintf('/bin/sh -c %s', escapeshellarg(ZID_PACKAGES_WATCHDOG_CMD . ' >/dev/null 2>&1'));
}

function zid_packages_install_watchdog_cron() {
	$cmd = zid_packages_watchdog_cron_command();
	config_lock();
	zid_packages_remove_cron_jobs_where(function ($job) use ($cmd) {
		$job_cmd = (string)($job['command'] ?? '');
		return $job_cmd !== '' && strpos($job_cmd, 'zid-packages watchdog') !== false;
	});
	zid_packages_upsert_cron_job(ZID_PACKAGES_WATCHDOG_DESCR, '*', '*', '*', '*', '*', 'root', $cmd);
	zid_packages_configure_cron_compat();
	@write_config('ZID Packages watchdog cron updated');
	config_unlock();
}

function zid_packages_remove_watchdog_cron() {
	config_lock();
	$changed = zid_packages_remove_cron_jobs_where(function ($job) {
		$job_cmd = (string)($job['command'] ?? '');
		$job_descr = (string)($job['descr'] ?? '');
		if ($job_descr === ZID_PACKAGES_WATCHDOG_DESCR) {
			return true;
		}
		if (strpos($job_cmd, 'zid-packages watchdog') !== false) {
			return true;
		}
		return false;
	});
	if ($changed) {
		zid_packages_configure_cron_compat();
		@write_config('ZID Packages watchdog cron removed');
	}
	config_unlock();
	return $changed;
}

function zid_packages_remove_legacy_watchdogs() {
	config_lock();
	$changed = zid_packages_remove_cron_jobs_where(function ($job) {
		$job_cmd = (string)($job['command'] ?? '');
		$job_descr = (string)($job['descr'] ?? '');
		if (strpos($job_cmd, 'zid-proxy-watchdog') !== false) {
			return true;
		}
		if (strpos($job_cmd, 'zid-geolocation-watchdog') !== false) {
			return true;
		}
		if (strpos($job_descr, 'ZID Proxy: watchdog') !== false) {
			return true;
		}
		if (strpos($job_descr, 'ZID Geolocation: watchdog') !== false) {
			return true;
		}
		return false;
	});
	if ($changed) {
		zid_packages_configure_cron_compat();
		@write_config('ZID Packages legacy watchdogs removed');
	}
	config_unlock();
	return $changed;
}

function zid_packages_remove_legacy_auto_update_cron() {
	config_lock();
	$changed = zid_packages_remove_cron_jobs_where(function ($job) {
		$job_cmd = (string)($job['command'] ?? '');
		$job_descr = (string)($job['descr'] ?? '');
		if (strpos($job_cmd, 'zid-packages auto-update') !== false) {
			return true;
		}
		if (strpos($job_descr, 'auto-update') !== false) {
			return true;
		}
		return false;
	});
	if ($changed) {
		zid_packages_configure_cron_compat();
		@write_config('ZID Packages legacy auto-update cron removed');
	}
	config_unlock();
	return $changed;
}

function zid_packages_log_tail($lines = 200) {
	if (!file_exists(ZID_PACKAGES_LOG)) {
		return array();
	}
	$lines = intval($lines);
	if ($lines <= 0) {
		$lines = 200;
	}
	$out = array();
	exec('/usr/bin/tail -n ' . $lines . ' ' . escapeshellarg(ZID_PACKAGES_LOG) . ' 2>/dev/null', $out);
	return $out;
}

function zid_packages_tabs($active) {
	$tabs = array();
	$tabs[] = array(gettext('Packages'), $active === 'packages', '/zid-packages_packages.php');
	$tabs[] = array(gettext('Services'), $active === 'services', '/zid-packages_services.php');
	$tabs[] = array(gettext('Licensing'), $active === 'licensing', '/zid-packages_licensing.php');
	$tabs[] = array(gettext('Logs'), $active === 'logs', '/zid-packages_logs.php');
	return $tabs;
}

?>
