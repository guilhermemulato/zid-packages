<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Packages"));
$pglinks = array("", "/zid-packages_packages.php", "@self");

$action_msg = '';
$action_output = '';
$action_status = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
	$pkg = $_POST['pkg'] ?? '';
	if (isset($_POST['install'])) {
		$cmd = escapeshellcmd(ZID_PACKAGES_BIN) . ' package install ' . escapeshellarg($pkg) . ' 2>&1';
		$out = array();
		$rc = 0;
		exec($cmd, $out, $rc);
		$action_output = trim(implode("\n", $out));
		$action_status = $rc === 0 ? 'success' : 'warning';
		$action_msg = $rc === 0 ? gettext('Install completed.') : gettext('Install failed.');
	}
	if (isset($_POST['update'])) {
		$cmd = escapeshellcmd(ZID_PACKAGES_BIN) . ' package update ' . escapeshellarg($pkg) . ' 2>&1';
		$out = array();
		$rc = 0;
		exec($cmd, $out, $rc);
		$action_output = trim(implode("\n", $out));
		$action_status = $rc === 0 ? 'success' : 'warning';
		$action_msg = $rc === 0 ? gettext('Update completed.') : gettext('Update failed.');
	}
}

$status_error = '';
$status = zid_packages_status_json($status_error);

include("head.inc");

display_top_tabs(zid_packages_tabs('packages'));

if ($status_error) {
	print_info_box(htmlspecialchars($status_error), 'warning');
}
if ($action_msg) {
	print_info_box(htmlspecialchars($action_msg), $action_status ?: 'info');
}
if ($action_output) {
	print_info_box('<pre>' . htmlspecialchars($action_output) . '</pre>', 'info');
}

$packages = $status['packages'] ?? array();
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Packages'); ?></h2></div>
	<div class="panel-body">
		<div class="table-responsive">
			<table class="table table-striped table-condensed">
				<thead>
					<tr>
						<th><?php echo gettext('Package'); ?></th>
						<th><?php echo gettext('Installed'); ?></th>
						<th><?php echo gettext('Version'); ?></th>
						<th><?php echo gettext('Remote'); ?></th>
						<th><?php echo gettext('Update'); ?></th>
						<th><?php echo gettext('Auto Update'); ?></th>
						<th><?php echo gettext('Service'); ?></th>
						<th><?php echo gettext('Enabled'); ?></th>
						<th><?php echo gettext('Licensed'); ?></th>
						<th><?php echo gettext('Action'); ?></th>
					</tr>
				</thead>
				<tbody>
				<?php foreach ($packages as $pkg): ?>
					<?php
						$key = $pkg['key'] ?? '';
						$is_self = $key === 'zid-packages';
						$installed = !empty($pkg['installed']);
						$remote_version = $pkg['version_remote'] ?? '';
						$local_version = $pkg['version_installed'] ?? '';
						$has_remote = $remote_version !== '';
						$update_available = !empty($pkg['update_available']);
						$running = !empty($pkg['service_running']);
						$enabled = !empty($pkg['enabled']);
						$licensed = !empty($pkg['licensed']);
						$up_to_date = $has_remote && $local_version !== '' && $remote_version === $local_version;
						$auto_age = intval($pkg['auto_update_age_days'] ?? 0);
						$auto_threshold = intval($pkg['auto_update_threshold_days'] ?? 0);
						$auto_due = !empty($pkg['auto_update_due']);
						$auto_due_at = intval($pkg['auto_update_due_at'] ?? 0);
					?>
					<tr>
						<td><?php echo htmlspecialchars($key); ?></td>
						<td><?php echo $installed ? gettext('Yes') : gettext('No'); ?></td>
						<td><?php echo htmlspecialchars($pkg['version_installed'] ?? '-'); ?></td>
						<td><?php echo $has_remote ? htmlspecialchars($remote_version) : '-'; ?></td>
						<td><?php echo $update_available ? gettext('Available') : ($has_remote ? gettext('OK') : gettext('N/A')); ?></td>
						<td>
							<?php if ($update_available): ?>
								<?php
									$auto_eta = $auto_due_at > 0 ? date('Y-m-d H:i', $auto_due_at) : '';
									if ($auto_due) {
										$auto_label = gettext('Due');
										if ($auto_eta !== '') {
											$auto_label .= ' (ETA ' . $auto_eta . ')';
										}
										echo htmlspecialchars($auto_label);
									} else {
										$auto_label = sprintf('%d/%d day', $auto_age, $auto_threshold);
										if ($auto_eta !== '') {
											$auto_label .= ' (ETA ' . $auto_eta . ')';
										}
										echo htmlspecialchars($auto_label);
									}
								?>
							<?php else: ?>
								-
							<?php endif; ?>
						</td>
						<td><?php echo $running ? gettext('Running') : gettext('Stopped'); ?></td>
						<td><?php echo $enabled ? gettext('Yes') : gettext('No'); ?></td>
						<td><?php echo $is_self ? gettext('N/A') : ($licensed ? gettext('Yes') : gettext('No')); ?></td>
						<td>
							<form method="post" style="margin:0">
								<input type="hidden" name="pkg" value="<?php echo htmlspecialchars($key); ?>" />
								<?php if (!$installed): ?>
									<button type="submit" name="install" class="btn btn-xs btn-success" onclick="return confirm('Install package now?');">
										<?php echo gettext('Install'); ?>
									</button>
								<?php else: ?>
									<?php if ($update_available): ?>
										<button type="submit" name="update" class="btn btn-xs btn-default" onclick="return confirm('Run update now?');">
											<?php echo gettext('Update'); ?>
										</button>
									<?php elseif ($up_to_date): ?>
										<span class="label label-success"><?php echo gettext('Up to date'); ?></span>
									<?php else: ?>
										<span class="label label-default"><?php echo gettext('N/A'); ?></span>
									<?php endif; ?>
								<?php endif; ?>
							</form>
						</td>
					</tr>
				<?php endforeach; ?>
				</tbody>
			</table>
		</div>
	</div>
</div>

<?php include("foot.inc"); ?>
