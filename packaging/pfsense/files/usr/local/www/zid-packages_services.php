<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Services"));
$pglinks = array("", "/zid-packages_services.php", "@self");

$status_error = '';
$status = zid_packages_status_json($status_error);
$services = $status['services'] ?? array();

include("head.inc");

display_top_tabs(zid_packages_tabs('services'));

if ($status_error) {
	print_info_box(htmlspecialchars($status_error), 'warning');
}
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Services'); ?></h2></div>
	<div class="panel-body">
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
