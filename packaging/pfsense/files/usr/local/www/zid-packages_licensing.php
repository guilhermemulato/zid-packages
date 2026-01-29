<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Licensing"));
$pglinks = array("", "/zid-packages_licensing.php", "@self");

$action_msg = '';
$action_output = '';
$action_status = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['license_sync'])) {
	$cmd = escapeshellcmd(ZID_PACKAGES_BIN) . ' license sync 2>&1';
	$out = array();
	$rc = 0;
	exec($cmd, $out, $rc);
	$action_output = trim(implode("\n", $out));
	$action_status = $rc === 0 ? 'success' : 'warning';
	$action_msg = $rc === 0 ? gettext('License updated.') : gettext('Failed to update license.');
}

$status_error = '';
$status = zid_packages_status_json($status_error);
$lic = $status['licensing'] ?? array();
$packages = $status['packages'] ?? array();

include("head.inc");

display_top_tabs(zid_packages_tabs('licensing'));

if ($status_error) {
	print_info_box(htmlspecialchars($status_error), 'warning');
}
if ($action_msg) {
	print_info_box(htmlspecialchars($action_msg), $action_status ?: 'info');
}
if ($action_output) {
	print_info_box('<pre>' . htmlspecialchars($action_output) . '</pre>', 'info');
}

function zid_packages_format_ts($ts) {
	$ts = intval($ts);
	if ($ts <= 0) {
		return '-';
	}
		return date('Y-m-d H:i:s', $ts);
}
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Licensing'); ?></h2></div>
	<div class="panel-body">
		<form method="post">
			<button type="submit" name="license_sync" class="btn btn-sm btn-primary" onclick="return confirm('Force license sync now?');">
				<?php echo gettext('Force sync'); ?>
			</button>
		</form>
		<br />
		<div class="table-responsive">
			<table class="table table-striped table-condensed">
				<tbody>
					<tr><th><?php echo gettext('Mode'); ?></th><td><?php echo htmlspecialchars($lic['mode'] ?? '-'); ?></td></tr>
					<tr><th><?php echo gettext('Reason'); ?></th><td><?php echo htmlspecialchars($lic['reason'] ?? '-'); ?></td></tr>
					<tr><th><?php echo gettext('Last attempt'); ?></th><td><?php echo zid_packages_format_ts($lic['last_attempt'] ?? 0); ?></td></tr>
					<tr><th><?php echo gettext('Last success'); ?></th><td><?php echo zid_packages_format_ts($lic['last_success'] ?? 0); ?></td></tr>
					<tr><th><?php echo gettext('Valid until'); ?></th><td><?php echo zid_packages_format_ts($lic['valid_until'] ?? 0); ?></td></tr>
				</tbody>
			</table>
		</div>
	</div>
</div>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Packages'); ?></h2></div>
	<div class="panel-body">
		<div class="table-responsive">
			<table class="table table-striped table-condensed">
				<thead>
					<tr>
						<th><?php echo gettext('Package'); ?></th>
						<th><?php echo gettext('Licensed'); ?></th>
					</tr>
				</thead>
				<tbody>
				<?php foreach ($packages as $pkg): ?>
					<?php $is_self = ($pkg['key'] ?? '') === 'zid-packages'; ?>
					<tr>
						<td><?php echo htmlspecialchars($pkg['key'] ?? ''); ?></td>
						<td><?php echo $is_self ? gettext('N/A') : (!empty($pkg['licensed']) ? gettext('Yes') : gettext('No')); ?></td>
					</tr>
				<?php endforeach; ?>
				</tbody>
			</table>
		</div>
	</div>
</div>

<?php include("foot.inc"); ?>
