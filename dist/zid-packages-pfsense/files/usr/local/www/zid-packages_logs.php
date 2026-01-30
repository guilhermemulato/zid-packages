<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Logs"));
$pglinks = array("", "/zid-packages_logs.php", "@self");

$lines = zid_packages_log_tail(50);
$lines = array_reverse($lines);
$auto = !empty($_GET['autorefresh']);

include("head.inc");

display_top_tabs(zid_packages_tabs('logs'));
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Logs'); ?></h2></div>
	<div class="panel-body">
		<form method="get" class="form-inline">
			<div class="form-group">
				<div class="btn-group" role="group" aria-label="Logs actions">
					<button type="submit" class="btn btn-xs btn-primary"><?php echo gettext('Refresh'); ?></button>
					<button type="button" class="btn btn-xs btn-default" onclick="var c=document.getElementById('autorefresh'); c.checked=!c.checked; window.location='?autorefresh=' + (c.checked ? '1' : '0');">
						<?php echo $auto ? gettext('Auto refresh: On') : gettext('Auto refresh: Off'); ?>
					</button>
				</div>
				<?php if ($auto): ?>
					<span class="label label-success" style="margin-left:6px;"><?php echo gettext('Auto'); ?></span>
				<?php else: ?>
					<span class="label label-default" style="margin-left:6px;"><?php echo gettext('Manual'); ?></span>
				<?php endif; ?>
				<input type="checkbox" id="autorefresh" style="display:none" <?php echo $auto ? 'checked' : ''; ?>>
			</div>
		</form>
		<br />
		<pre><?php echo htmlspecialchars(implode("\n", $lines)); ?></pre>
	</div>
</div>

<?php if ($auto): ?>
<script>
setTimeout(function () { window.location.reload(); }, 5000);
</script>
<?php endif; ?>

<?php include("foot.inc"); ?>
