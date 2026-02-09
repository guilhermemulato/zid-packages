<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Packages"));
$pglinks = array("", "/zid-packages_packages.php", "@self");

$action_msg = '';
$action_output = '';
$action_status = '';

if (isset($_GET['ajax'])) {
	$action = $_POST['action'] ?? $_GET['action'] ?? '';
	$pkg = $_POST['pkg'] ?? $_GET['pkg'] ?? '';
	$resp = array('ok' => false);
	if ($action === 'start_update') {
		$error = '';
		$ok = zid_packages_start_update_bg($pkg, $error);
		$resp['ok'] = $ok;
		if (!$ok) {
			$resp['error'] = $error;
		}
		$resp['status'] = zid_packages_update_status_read($pkg);
	} elseif ($action === 'status') {
		$status = zid_packages_update_status_read($pkg);
		$resp['ok'] = true;
		$resp['status'] = $status;
		$log_path = $status['log'] ?? '';
		$offset = intval($_POST['offset'] ?? $_GET['offset'] ?? 0);
		$log_info = zid_packages_log_read_from($log_path, $offset, 65536);
		$resp['log_chunk'] = $log_info['data'];
		$resp['log_offset'] = $log_info['offset'];
		$resp['log_truncated'] = $log_info['truncated'];
	} else {
		$resp['error'] = 'Invalid action.';
	}
	header('Content-Type: application/json');
	echo json_encode($resp);
	exit;
}

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
		$error = '';
		$ok = zid_packages_start_update_bg($pkg, $error);
		$action_output = '';
		$action_status = $ok ? 'info' : 'warning';
		$action_msg = $ok ? gettext('Update started in background. You can follow progress below.') : gettext('Failed to start update.');
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
						$update_state = zid_packages_update_status_read($key);
						$update_running = !empty($update_state['running']);
						$update_started_at = intval($update_state['started_at'] ?? 0);
						$update_finished_at = intval($update_state['finished_at'] ?? 0);
						$update_exit_code = $update_state['exit_code'] ?? null;
						$update_pid = intval($update_state['pid'] ?? 0);
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
								<?php if (function_exists('csrfguard_form')) { csrfguard_form(); } ?>
								<input type="hidden" name="pkg" value="<?php echo htmlspecialchars($key); ?>" />
								<?php if (!$installed): ?>
									<button type="submit" name="install" class="btn btn-xs btn-success" onclick="return confirm('Install package now?');">
										<?php echo gettext('Install'); ?>
									</button>
								<?php else: ?>
									<?php if ($update_available): ?>
										<button type="submit" name="update" class="btn btn-xs btn-default js-update-btn" data-pkg="<?php echo htmlspecialchars($key); ?>" onclick="return confirm('Run update now?');">
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
					<tr class="update-log-row"
						data-pkg="<?php echo htmlspecialchars($key); ?>"
						data-running="<?php echo $update_running ? '1' : '0'; ?>"
						data-started="<?php echo htmlspecialchars((string)$update_started_at); ?>"
						data-finished="<?php echo htmlspecialchars((string)$update_finished_at); ?>"
						data-exit="<?php echo htmlspecialchars((string)($update_exit_code ?? '')); ?>"
						data-pid="<?php echo htmlspecialchars((string)$update_pid); ?>"
						style="display:none">
						<td colspan="10">
							<div class="update-status text-muted">-</div>
							<div class="update-actions" style="margin-top:6px; display:none;">
								<button type="button" class="btn btn-xs btn-default update-close">
									<?php echo gettext('Close'); ?>
								</button>
								<button type="button" class="btn btn-xs btn-default update-reload" style="margin-left:6px; display:none;">
									<?php echo gettext('Reload page'); ?>
								</button>
							</div>
							<pre class="update-log" style="display:none; max-height:200px; overflow:auto; margin-top:6px;"></pre>
						</td>
					</tr>
				<?php endforeach; ?>
				</tbody>
			</table>
		</div>
	</div>
</div>

<script type="text/javascript">
(function() {
	if (!window.fetch) {
		return;
	}

	var pollers = {};
	var pollIntervalMs = 2000;
	var statusRowByPkg = {};
	var logState = {};

	function $(selector, root) {
		return (root || document).querySelector(selector);
	}

	function $all(selector, root) {
		return Array.prototype.slice.call((root || document).querySelectorAll(selector));
	}

	function encodeForm(data) {
		if (window.URLSearchParams) {
			var params = new URLSearchParams();
			Object.keys(data || {}).forEach(function(key) {
				if (data[key] !== undefined && data[key] !== null) {
					params.append(key, data[key]);
				}
			});
			return params.toString();
		}
		var pairs = [];
		Object.keys(data || {}).forEach(function(key) {
			if (data[key] !== undefined && data[key] !== null) {
				pairs.push(encodeURIComponent(key) + '=' + encodeURIComponent(data[key]));
			}
		});
		return pairs.join('&');
	}

	function formatTs(ts) {
		if (!ts || ts <= 0) {
			return '-';
		}
		try {
			return new Date(ts * 1000).toLocaleString();
		} catch (e) {
			return String(ts);
		}
	}

	function ensureRow(pkg) {
		if (statusRowByPkg[pkg]) {
			return statusRowByPkg[pkg];
		}
		var row = $('.update-log-row[data-pkg="' + pkg + '"]');
		if (!row) {
			return null;
		}
		statusRowByPkg[pkg] = row;
		return row;
	}

	function showRow(pkg) {
		var row = ensureRow(pkg);
		if (!row) {
			return;
		}
		row.style.display = '';
		var actions = $('.update-actions', row);
		if (actions) {
			actions.style.display = 'block';
		}
		if (!logState[pkg]) {
			logState[pkg] = { offset: 0, runKey: '', visible: true };
		} else {
			logState[pkg].visible = true;
		}
	}

	function hideRow(pkg) {
		var row = ensureRow(pkg);
		if (!row) {
			return;
		}
		row.style.display = 'none';
		if (pollers[pkg]) {
			clearInterval(pollers[pkg]);
			delete pollers[pkg];
		}
		if (!logState[pkg]) {
			logState[pkg] = { offset: 0, runKey: '', visible: false };
		} else {
			logState[pkg].visible = false;
		}
	}

	function getRunKey(status) {
		if (!status) {
			return '';
		}
		return String(status.started_at || 0) + ':' + String(status.pid || 0);
	}

	function setRow(pkg, status) {
		var row = ensureRow(pkg);
		if (!row) {
			return;
		}
		var statusEl = $('.update-status', row);
		var logEl = $('.update-log', row);
		var reloadBtn = $('.update-reload', row);
		var closeBtn = $('.update-close', row);
		var running = !!(status && status.running);
		var exitCode = status ? status.exit_code : null;
		var startedAt = status ? status.started_at : null;
		var finishedAt = status ? status.finished_at : null;
		var msg = '-';
		if (running) {
			msg = '<?php echo gettext('Running'); ?>' + ' (start ' + formatTs(startedAt) + ')';
		} else if (exitCode !== null && exitCode !== undefined) {
			if (exitCode === 0 || exitCode === '0') {
				msg = '<?php echo gettext('Completed'); ?>' + ' (finish ' + formatTs(finishedAt) + ')';
			} else {
				msg = '<?php echo gettext('Failed'); ?>' + ' (code ' + exitCode + ', finish ' + formatTs(finishedAt) + ')';
			}
		}
		statusEl.textContent = msg;
		if (closeBtn) {
			closeBtn.style.display = running ? 'none' : 'inline-block';
		}
		if (!running && (exitCode !== null && exitCode !== undefined) && reloadBtn) {
			reloadBtn.style.display = 'inline-block';
		}
	}

	function parseJsonResponse(resp) {
		return resp.text().then(function(text) {
			try {
				return JSON.parse(text);
			} catch (e) {
				return { ok: false, error: 'Invalid response from server.' };
			}
		});
	}

	function ajaxPost(data) {
		return fetch('zid-packages_packages.php?ajax=1', {
			method: 'POST',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
			body: encodeForm(data || {}),
			credentials: 'same-origin',
			cache: 'no-store'
		}).then(parseJsonResponse);
	}

	function ajaxStatus(pkg, offset) {
		var qs = encodeForm({ action: 'status', pkg: pkg, offset: offset || 0, ts: Date.now() });
		return fetch('zid-packages_packages.php?ajax=1&' + qs, {
			method: 'GET',
			credentials: 'same-origin',
			cache: 'no-store'
		}).then(parseJsonResponse);
	}

	function pollStatus(pkg) {
		var state = logState[pkg] || { offset: 0, runKey: '', visible: false };
		return ajaxStatus(pkg, state.offset).then(function(data) {
			if (!data || !data.ok) {
				return;
			}
			var status = data.status || {};
			var runKey = getRunKey(status);
			if (status.running && !state.visible) {
				showRow(pkg);
			}
			if (runKey && state.runKey && state.runKey !== runKey) {
				var row = ensureRow(pkg);
				if (row) {
					var logEl = $('.update-log', row);
					logEl.textContent = '';
					logEl.style.display = 'none';
				}
				state.offset = 0;
			}
			state.runKey = runKey;
			var chunk = data.log_chunk || '';
			var newOffset = data.log_offset;
			if (chunk && chunk.length > 0) {
				var row = ensureRow(pkg);
				if (row) {
					var logEl = $('.update-log', row);
					logEl.textContent += chunk;
					logEl.style.display = 'block';
					showRow(pkg);
				}
			}
			if (newOffset !== undefined && newOffset !== null) {
				state.offset = newOffset;
			}
			logState[pkg] = state;
			if (state.visible) {
				setRow(pkg, status);
				if (!status.running && (status.exit_code !== null && status.exit_code !== undefined)) {
					// keep visible until user closes
				} else if (!status.running && (status.exit_code === null || status.exit_code === undefined)) {
					hideRow(pkg);
				}
			}
			if (status.running) {
				if (!pollers[pkg]) {
					pollers[pkg] = setInterval(function() { pollStatus(pkg); }, pollIntervalMs);
				}
			} else if (pollers[pkg]) {
				clearInterval(pollers[pkg]);
				delete pollers[pkg];
			}
		});
	}

	function startUpdate(pkg, formEl) {
		var csrf = '';
		if (formEl) {
			var csrfEl = formEl.querySelector('input[name="__csrf_magic"]');
			if (csrfEl && csrfEl.value) {
				csrf = csrfEl.value;
			}
		}
		ajaxPost({ action: 'start_update', pkg: pkg, __csrf_magic: csrf }).then(function(data) {
			if (!data || !data.ok) {
				if (formEl) {
					formEl.submit();
					return;
				}
				alert((data && data.error) ? data.error : 'Failed to start update.');
				return;
			}
			var status = data.status || {};
			logState[pkg] = { offset: 0, runKey: getRunKey(status), visible: true };
			var row = ensureRow(pkg);
			if (row) {
				var logEl = $('.update-log', row);
				logEl.textContent = '';
				logEl.style.display = 'none';
			}
			showRow(pkg);
			setRow(pkg, status);
			pollStatus(pkg);
		}).catch(function() {
			if (formEl) {
				formEl.submit();
				return;
			}
			alert('Failed to start update.');
		});
	}

	$all('.js-update-btn').forEach(function(btn) {
		btn.removeAttribute('onclick');
		btn.addEventListener('click', function(ev) {
			if (!window.fetch) {
				return;
			}
			if (!confirm('Run update now?')) {
				ev.preventDefault();
				return;
			}
			ev.preventDefault();
			var pkg = btn.getAttribute('data-pkg') || '';
			if (!pkg) {
				return;
			}
			showRow(pkg);
			startUpdate(pkg, btn.form);
		});
	});

	$all('.update-log-row').forEach(function(row) {
		var pkg = row.getAttribute('data-pkg');
		if (pkg) {
			logState[pkg] = logState[pkg] || { offset: 0, runKey: '', visible: false };
			var running = row.getAttribute('data-running') === '1';
			var started = parseInt(row.getAttribute('data-started') || '0', 10) || 0;
			var finished = parseInt(row.getAttribute('data-finished') || '0', 10) || 0;
			var exitCode = row.getAttribute('data-exit');
			var pid = parseInt(row.getAttribute('data-pid') || '0', 10) || 0;
			if (running) {
				logState[pkg].visible = true;
				showRow(pkg);
				setRow(pkg, {
					running: true,
					started_at: started,
					finished_at: finished,
					exit_code: exitCode !== '' ? exitCode : null,
					pid: pid
				});
				pollStatus(pkg);
			}
		}
	});

	$all('.update-close').forEach(function(btn) {
		btn.addEventListener('click', function() {
			var row = btn.closest('.update-log-row');
			if (!row) {
				return;
			}
			var pkg = row.getAttribute('data-pkg') || '';
			if (pkg) {
				hideRow(pkg);
			}
		});
	});

	$all('.update-reload').forEach(function(btn) {
		btn.addEventListener('click', function() {
			window.location.reload();
		});
	});
})();
</script>

<?php include("foot.inc"); ?>
