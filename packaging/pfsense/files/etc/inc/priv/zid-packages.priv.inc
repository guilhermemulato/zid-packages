<?php
/*
 * zid-packages.priv.inc
 */

global $priv_list;

$priv_list['page-services-zidpackages'] = array();
$priv_list['page-services-zidpackages']['name'] = 'WebCfg - Services: ZID Packages';
$priv_list['page-services-zidpackages']['descr'] = 'Allow access to ZID Packages manager';
$priv_list['page-services-zidpackages']['match'] = array();
$priv_list['page-services-zidpackages']['match'][] = 'pkg.php?xml=zid-packages.xml*';
$priv_list['page-services-zidpackages']['match'][] = 'pkg_edit.php?xml=zid-packages.xml*';
$priv_list['page-services-zidpackages']['match'][] = 'zid-packages_packages.php*';
$priv_list['page-services-zidpackages']['match'][] = 'zid-packages_services.php*';
$priv_list['page-services-zidpackages']['match'][] = 'zid-packages_licensing.php*';
$priv_list['page-services-zidpackages']['match'][] = 'zid-packages_logs.php*';

?>
