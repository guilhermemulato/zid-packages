<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Services"));
$pglinks = array("", "/zid-packages_services.php", "@self");

$action_msg = '';
$action_output = '';
$action_status = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['zid_packages_action'])) {
	$action = $_POST['zid_packages_action'];
	$rc = 0;
	$output = '';
	$rc_cmd = '/usr/local/etc/rc.d/zid_packages';
	if ($action === 'start') {
		zid_packages_run_cmd(escapeshellcmd($rc_cmd) . ' onestart 2>&1', $output, $rc);
		$action_msg = $rc === 0 ? gettext('ZID Packages started.') : gettext('Failed to start ZID Packages.');
	} elseif ($action === 'stop') {
		zid_packages_run_cmd(escapeshellcmd($rc_cmd) . ' onestop 2>&1', $output, $rc);
		$action_msg = $rc === 0 ? gettext('ZID Packages stopped.') : gettext('Failed to stop ZID Packages.');
	} elseif ($action === 'restart') {
		zid_packages_run_cmd(escapeshellcmd($rc_cmd) . ' onerestart 2>&1', $output, $rc);
		$action_msg = $rc === 0 ? gettext('ZID Packages restarted.') : gettext('Failed to restart ZID Packages.');
	}
	$action_output = trim($output);
	$action_status = $rc === 0 ? 'success' : 'warning';
}

$status_error = '';
$status = zid_packages_status_json($status_error);
$services = $status['services'] ?? array();
$zid_running = false;
if (isset($status['services']) && is_array($status['services'])) {
	foreach ($status['services'] as $svc) {
		if (($svc['key'] ?? '') === 'zid-packages') {
			$zid_running = !empty($svc['running']);
			break;
		}
	}
}
if (!$zid_running && isset($status['packages']) && is_array($status['packages'])) {
	foreach ($status['packages'] as $pkg) {
		if (($pkg['key'] ?? '') === 'zid-packages') {
			$zid_running = !empty($pkg['service_running']);
			break;
		}
	}
}

include("head.inc");

display_top_tabs(zid_packages_tabs('services'));

if ($status_error) {
	print_info_box(htmlspecialchars($status_error), 'warning');
}
if ($action_msg) {
	print_info_box(htmlspecialchars($action_msg), $action_status ?: 'info');
}
if ($action_output) {
	print_info_box('<pre>' . htmlspecialchars($action_output) . '</pre>', 'info');
}
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Services'); ?></h2></div>
	<div class="panel-body">
		<form method="post" class="form-inline">
			<div class="form-group">
				<?php if ($zid_running): ?>
					<button type="submit" name="zid_packages_action" value="restart" class="btn btn-xs btn-warning"><?php echo gettext('Restart ZID Packages'); ?></button>
					<button type="submit" name="zid_packages_action" value="stop" class="btn btn-xs btn-danger" onclick="return confirm('Stop ZID Packages?');"><?php echo gettext('Stop ZID Packages'); ?></button>
				<?php else: ?>
					<button type="submit" name="zid_packages_action" value="start" class="btn btn-xs btn-success"><?php echo gettext('Start ZID Packages'); ?></button>
				<?php endif; ?>
			</div>
		</form>
		<br />
		<div class="table-responsive">
			<table class="table table-striped table-condensed">
				<thead>
					<tr>
						<th><?php echo gettext('Service'); ?></th>
						<th><?php echo gettext('Running'); ?></th>
						<th><?php echo gettext('Enabled'); ?></th>
						<th><?php echo gettext('Licensed'); ?></th>
					</tr>
				</thead>
				<tbody>
				<?php foreach ($services as $svc): ?>
					<tr>
						<td><?php echo htmlspecialchars($svc['name'] ?? $svc['key']); ?></td>
						<td><?php echo !empty($svc['running']) ? gettext('Running') : gettext('Stopped'); ?></td>
						<td><?php echo !empty($svc['enabled']) ? gettext('Yes') : gettext('No'); ?></td>
						<td><?php echo !empty($svc['licensed']) ? gettext('Yes') : gettext('No'); ?></td>
					</tr>
				<?php endforeach; ?>
				</tbody>
			</table>
		</div>
	</div>
</div>

<?php include("foot.inc"); ?>
