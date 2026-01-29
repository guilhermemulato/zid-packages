<?php
require_once("guiconfig.inc");
require_once("/usr/local/pkg/zid-packages.inc");

$pgtitle = array(gettext("Services"), gettext("ZID Packages"), gettext("Logs"));
$pglinks = array("", "/zid-packages_logs.php", "@self");

$lines = zid_packages_log_tail(200);

include("head.inc");

display_top_tabs(zid_packages_tabs('logs'));
?>

<div class="panel panel-default">
	<div class="panel-heading"><h2 class="panel-title"><?php echo gettext('Logs'); ?></h2></div>
	<div class="panel-body">
		<pre><?php echo htmlspecialchars(implode("\n", $lines)); ?></pre>
	</div>
</div>

<?php include("foot.inc"); ?>
