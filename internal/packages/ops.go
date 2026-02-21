package packages

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"zid-packages/internal/logx"
	"zid-packages/internal/s3"
)

const (
	proxyBin        = "/usr/local/sbin/zid-proxy"
	appidBin        = "/usr/local/sbin/zid-appid"
	threatdBin      = "/usr/local/sbin/zid-threatd"
	geolocationBin  = "/usr/local/sbin/zid-geolocation"
	logsBin         = "/usr/local/sbin/zid-logs"
	accessBin       = "/usr/local/sbin/zidaccess"
	orchestratorBin = "/usr/local/sbin/zid-orchestration"
	packagesBin     = "/usr/local/sbin/zid-packages"
)

const enabledCacheTTL = 2 * time.Minute

type enabledCacheEntry struct {
	value     bool
	timestamp time.Time
}

var (
	enabledCacheMu sync.Mutex
	enabledCache   = map[string]enabledCacheEntry{}
	enableDebug    = envTrue("ZID_PACKAGES_ENABLE_DEBUG")
	enableLogger   = logx.New("/var/log/zid-packages.log")
)

func Installed(key string) bool {
	switch key {
	case "zid-packages":
		return fileExists(packagesBin)
	case "zid-proxy":
		return fileExists(proxyBin)
	case "zid-geolocation":
		return fileExists(geolocationBin)
	case "zid-logs":
		return fileExists(logsBin)
	case "zid-access":
		return fileExists(accessBin)
	case "zid-orchestrator":
		return fileExists(orchestratorBin)
	default:
		return false
	}
}

func Enabled(key string) (bool, error) {
	switch key {
	case "zid-packages":
		if b, ok := readRCConfBool("/etc/rc.conf.local", "zid_packages_enable"); ok {
			return b, nil
		}
		if b, ok := readRCConfBool("/etc/rc.conf", "zid_packages_enable"); ok {
			return b, nil
		}
		return false, nil
	case "zid-proxy":
		if b, ok := readEnableViaPHP("zid-proxy"); ok {
			logEnable(key, "php:installedpackages/zidproxy/config/enable", boolString(b), true)
			return cacheEnabled(key, b), nil
		}
		val, ok := readConfigXMLValueRetry([]string{"installedpackages", "zidproxy", "config", "enable"}, 3)
		logEnable(key, "config:installedpackages/zidproxy/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueRetry([]string{"zidproxy", "config", "enable"}, 3)
		logEnable(key, "config:zidproxy/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"installedpackages", "zidproxy", "config", "enable"}, 3)
		logEnable(key, "config-loose:installedpackages/zidproxy/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"zidproxy", "config", "enable"}, 3)
		logEnable(key, "config-loose:zidproxy/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		if cached, ok := cachedEnabled(key); ok {
			logEnable(key, "cache", boolString(cached), true)
			return cached, nil
		}
		return false, nil
	case "zid-threatd":
		if b, ok := readEnableViaPHP("zid-threatd"); ok {
			logEnable(key, "php:installedpackages/zidproxy/config/threat_enable", boolString(b), true)
			return cacheEnabled(key, b), nil
		}
		val, ok := readConfigXMLValueRetry([]string{"installedpackages", "zidproxy", "config", "threat_enable"}, 3)
		logEnable(key, "config:installedpackages/zidproxy/config/threat_enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueRetry([]string{"zidproxy", "config", "threat_enable"}, 3)
		logEnable(key, "config:zidproxy/config/threat_enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"installedpackages", "zidproxy", "config", "threat_enable"}, 3)
		logEnable(key, "config-loose:installedpackages/zidproxy/config/threat_enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"zidproxy", "config", "threat_enable"}, 3)
		logEnable(key, "config-loose:zidproxy/config/threat_enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		if cached, ok := cachedEnabled(key); ok {
			logEnable(key, "cache", boolString(cached), true)
			return cached, nil
		}
		return false, nil
	case "zid-geolocation":
		if b, ok := readEnableViaPHP("zid-geolocation"); ok {
			logEnable(key, "php:installedpackages/zidgeolocation/config/enable", boolString(b), true)
			return cacheEnabled(key, b), nil
		}
		val, ok := readConfigXMLValueRetry([]string{"installedpackages", "zidgeolocation", "config", "enable"}, 3)
		logEnable(key, "config:installedpackages/zidgeolocation/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueRetry([]string{"zidgeolocation", "config", "enable"}, 3)
		logEnable(key, "config:zidgeolocation/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"installedpackages", "zidgeolocation", "config", "enable"}, 3)
		logEnable(key, "config-loose:installedpackages/zidgeolocation/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		val, ok = readConfigXMLValueLooseRetry([]string{"zidgeolocation", "config", "enable"}, 3)
		logEnable(key, "config-loose:zidgeolocation/config/enable", val, ok)
		if ok {
			return cacheEnabled(key, isOn(val)), nil
		}
		if b, ok := readJSONBool("/usr/local/etc/zid-geolocation/config.json", "enable"); ok {
			logEnable(key, "config-json:/usr/local/etc/zid-geolocation/config.json", boolString(b), true)
			return cacheEnabled(key, b), nil
		}
		if cached, ok := cachedEnabled(key); ok {
			logEnable(key, "cache", boolString(cached), true)
			return cached, nil
		}
		return false, nil
	case "zid-logs":
		if b, ok := readJSONBool("/usr/local/etc/zid-logs/config.json", "enabled"); ok {
			return b, nil
		}
		return false, nil
	case "zid-access":
		if b, ok := readEnableViaPHP("zid-access"); ok {
			logEnable(key, "php:installedpackages/(zidaccess|zid-access|zid_access)/config/enable", boolString(b), true)
			return cacheEnabled(key, b), nil
		}
		sections := []string{"zidaccess", "zid-access", "zid_access"}
		for _, section := range sections {
			val, ok := readConfigXMLValueRetry([]string{"installedpackages", section, "config", "enable"}, 3)
			logEnable(key, "config:installedpackages/"+section+"/config/enable", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
			val, ok = readConfigXMLValueRetry([]string{section, "config", "enable"}, 3)
			logEnable(key, "config:"+section+"/config/enable", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
		}
		// Formato quebrado (legado): lista escalar sem chaves (gera <config>valor</config> repetido).
		// Nesse caso o primeiro <config> costuma ser o enable.
		for _, section := range sections {
			val, ok := readConfigXMLValueRetry([]string{"installedpackages", section, "config"}, 3)
			logEnable(key, "config:installedpackages/"+section+"/config", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
			val, ok = readConfigXMLValueRetry([]string{section, "config"}, 3)
			logEnable(key, "config:"+section+"/config", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
		}
		for _, section := range sections {
			val, ok := readConfigXMLValueLooseRetry([]string{"installedpackages", section, "config", "enable"}, 3)
			logEnable(key, "config-loose:installedpackages/"+section+"/config/enable", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
			val, ok = readConfigXMLValueLooseRetry([]string{section, "config", "enable"}, 3)
			logEnable(key, "config-loose:"+section+"/config/enable", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
		}
		for _, section := range sections {
			val, ok := readConfigXMLValueLooseRetry([]string{"installedpackages", section, "config"}, 3)
			logEnable(key, "config-loose:installedpackages/"+section+"/config", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
			val, ok = readConfigXMLValueLooseRetry([]string{section, "config"}, 3)
			logEnable(key, "config-loose:"+section+"/config", val, ok)
			if ok {
				return cacheEnabled(key, isOn(val)), nil
			}
		}
		if cached, ok := cachedEnabled(key); ok {
			logEnable(key, "cache", boolString(cached), true)
			return cached, nil
		}
		return false, nil
	case "zid-orchestrator":
		if b, ok := readRCConfBool("/etc/rc.conf.local", "zid_orchestration_enable"); ok {
			logEnable(key, "rc.conf.local:zid_orchestration_enable", boolString(b), true)
			return b, nil
		}
		if b, ok := readRCConfBool("/etc/rc.conf", "zid_orchestration_enable"); ok {
			logEnable(key, "rc.conf:zid_orchestration_enable", boolString(b), true)
			return b, nil
		}
		logEnable(key, "rc.conf*:zid_orchestration_enable", "", false)
		return false, nil
	default:
		return false, errors.New("unknown package")
	}
}

func EnableSnapshot(key string) map[string]string {
	out := map[string]string{}
	switch key {
	case "zid-proxy":
		out["config:installedpackages/zidproxy/config/enable"] = readValueOrEmpty([]string{"installedpackages", "zidproxy", "config", "enable"})
		out["config:zidproxy/config/enable"] = readValueOrEmpty([]string{"zidproxy", "config", "enable"})
		out["config-loose:installedpackages/zidproxy/config/enable"] = readValueLooseOrEmpty([]string{"installedpackages", "zidproxy", "config", "enable"})
		out["config-loose:zidproxy/config/enable"] = readValueLooseOrEmpty([]string{"zidproxy", "config", "enable"})
	case "zid-threatd":
		out["config:installedpackages/zidproxy/config/threat_enable"] = readValueOrEmpty([]string{"installedpackages", "zidproxy", "config", "threat_enable"})
		out["config:zidproxy/config/threat_enable"] = readValueOrEmpty([]string{"zidproxy", "config", "threat_enable"})
		out["config-loose:installedpackages/zidproxy/config/threat_enable"] = readValueLooseOrEmpty([]string{"installedpackages", "zidproxy", "config", "threat_enable"})
		out["config-loose:zidproxy/config/threat_enable"] = readValueLooseOrEmpty([]string{"zidproxy", "config", "threat_enable"})
	case "zid-geolocation":
		out["config:installedpackages/zidgeolocation/config/enable"] = readValueOrEmpty([]string{"installedpackages", "zidgeolocation", "config", "enable"})
		out["config:zidgeolocation/config/enable"] = readValueOrEmpty([]string{"zidgeolocation", "config", "enable"})
		out["config-loose:installedpackages/zidgeolocation/config/enable"] = readValueLooseOrEmpty([]string{"installedpackages", "zidgeolocation", "config", "enable"})
		out["config-loose:zidgeolocation/config/enable"] = readValueLooseOrEmpty([]string{"zidgeolocation", "config", "enable"})
		if b, ok := readJSONBool("/usr/local/etc/zid-geolocation/config.json", "enable"); ok {
			out["config-json:/usr/local/etc/zid-geolocation/config.json"] = boolString(b)
		}
	case "zid-access":
		if b, ok := readEnableViaPHP("zid-access"); ok {
			out["php:installedpackages/(zidaccess|zid-access|zid_access)/config/enable"] = boolString(b)
		} else {
			out["php:installedpackages/(zidaccess|zid-access|zid_access)/config/enable"] = ""
		}
		sections := []string{"zidaccess", "zid-access", "zid_access"}
		for _, section := range sections {
			out["config:installedpackages/"+section+"/config/enable"] = readValueOrEmpty([]string{"installedpackages", section, "config", "enable"})
			out["config:"+section+"/config/enable"] = readValueOrEmpty([]string{section, "config", "enable"})
			out["config-loose:installedpackages/"+section+"/config/enable"] = readValueLooseOrEmpty([]string{"installedpackages", section, "config", "enable"})
			out["config-loose:"+section+"/config/enable"] = readValueLooseOrEmpty([]string{section, "config", "enable"})
			out["config:installedpackages/"+section+"/config"] = readValueOrEmpty([]string{"installedpackages", section, "config"})
			out["config:"+section+"/config"] = readValueOrEmpty([]string{section, "config"})
			out["config-loose:installedpackages/"+section+"/config"] = readValueLooseOrEmpty([]string{"installedpackages", section, "config"})
			out["config-loose:"+section+"/config"] = readValueLooseOrEmpty([]string{section, "config"})
		}
	case "zid-orchestrator":
		if b, ok := readRCConfBool("/etc/rc.conf.local", "zid_orchestration_enable"); ok {
			out["rc.conf.local:zid_orchestration_enable"] = boolString(b)
		} else {
			out["rc.conf.local:zid_orchestration_enable"] = ""
		}
		if b, ok := readRCConfBool("/etc/rc.conf", "zid_orchestration_enable"); ok {
			out["rc.conf:zid_orchestration_enable"] = boolString(b)
		} else {
			out["rc.conf:zid_orchestration_enable"] = ""
		}
	}
	return out
}

func ServiceRunning(key string) (bool, error) {
	switch key {
	case "zid-packages":
		return pgrepRunning("^/usr/local/sbin/zid-packages daemon"), nil
	case "zid-proxy":
		return pgrepRunning("^/usr/local/sbin/zid-proxy"), nil
	case "zid-geolocation":
		return pgrepRunning("^/usr/local/sbin/zid-geolocation"), nil
	case "zid-logs":
		return pgrepRunning("^/usr/local/sbin/zid-logs"), nil
	case "zid-access":
		return pgrepRunning("/usr/local/sbin/zidaccess"), nil
	case "zid-appid":
		return pgrepRunning("^/usr/local/sbin/zid-appid"), nil
	case "zid-threatd":
		return pgrepRunning("^/usr/local/sbin/zid-threatd"), nil
	case "zid-orchestrator":
		return pgrepRunning("^/usr/local/sbin/zid-orchestration"), nil
	default:
		return false, errors.New("unknown service")
	}
}

func StartService(key string) error {
	switch key {
	case "zid-packages":
		return run("/usr/local/etc/rc.d/zid_packages", "onestart")
	case "zid-proxy":
		return run("/usr/local/etc/rc.d/zid-proxy.sh", "start")
	case "zid-geolocation":
		if err := run("/usr/local/etc/rc.d/zid_geolocation", "onestart"); err != nil {
			return err
		}
		return runServiceStartPostAction(key)
	case "zid-logs":
		return run("/usr/local/etc/rc.d/zid_logs", "onestart")
	case "zid-access":
		return run("/usr/local/etc/rc.d/zid-access.sh", "start")
	case "zid-appid":
		return startAppID()
	case "zid-threatd":
		if !fileExists(threatdBin) {
			return errors.New("threatd binary not found")
		}
		return run("/usr/local/etc/rc.d/zid-threatd", "start")
	case "zid-orchestrator":
		return run("/usr/local/etc/rc.d/zid_orchestration", "onestart")
	default:
		return errors.New("unknown service")
	}
}

func StopService(key string) error {
	switch key {
	case "zid-packages":
		return run("/usr/local/etc/rc.d/zid_packages", "onestop")
	case "zid-proxy":
		stopErr := run("/usr/local/etc/rc.d/zid-proxy.sh", "stop")
		cleanupErr := runServiceFirewallCleanup(key)
		return errors.Join(stopErr, cleanupErr)
	case "zid-geolocation":
		stopErr := run("/usr/local/etc/rc.d/zid_geolocation", "onestop")
		cleanupErr := runServiceFirewallCleanup(key)
		return errors.Join(stopErr, cleanupErr)
	case "zid-logs":
		return run("/usr/local/etc/rc.d/zid_logs", "onestop")
	case "zid-access":
		return run("/usr/local/etc/rc.d/zid-access.sh", "stop")
	case "zid-appid":
		return stopAppID()
	case "zid-threatd":
		return run("/usr/local/etc/rc.d/zid-threatd", "stop")
	case "zid-orchestrator":
		return run("/usr/local/etc/rc.d/zid_orchestration", "onestop")
	default:
		return errors.New("unknown service")
	}
}

func runServiceStartPostAction(key string) error {
	includePath, functionName, ok := serviceStartPostAction(key)
	if !ok {
		return nil
	}
	if !fileExists(includePath) {
		return nil
	}
	php := phpBin()
	if php == "" {
		return nil
	}
	cmd := exec.Command(php, "-r", servicePHPFunctionScript(includePath, functionName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func serviceStartPostAction(key string) (includePath string, functionName string, ok bool) {
	switch key {
	case "zid-geolocation":
		return "/usr/local/pkg/zid-geolocation.inc", "zid_geolocation_apply_async", true
	default:
		return "", "", false
	}
}

func runServiceFirewallCleanup(key string) error {
	includePath, functionName, ok := serviceFirewallCleanupAction(key)
	if !ok {
		return nil
	}
	if !fileExists(includePath) {
		return nil
	}
	php := phpBin()
	if php == "" {
		return nil
	}
	cmd := exec.Command(php, "-r", servicePHPFunctionScript(includePath, functionName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func serviceFirewallCleanupAction(key string) (includePath string, functionName string, ok bool) {
	switch key {
	case "zid-proxy":
		return "/usr/local/pkg/zid-proxy.inc", "zidproxy_service_poststop_hook", true
	case "zid-geolocation":
		return "/usr/local/pkg/zid-geolocation.inc", "zid_geolocation_clear_floating_rules", true
	default:
		return "", "", false
	}
}

func servicePHPFunctionScript(includePath, functionName string) string {
	return `require_once("` + includePath + `"); if (function_exists("` + functionName + `")) { ` + functionName + `(); }`
}

func VersionLocal(key string) string {
	switch key {
	case "zid-packages":
		if v := readConfigXMLPackageVersion("zid-packages"); v != "" {
			return v
		}
		return readBinaryVersion(packagesBin)
	case "zid-proxy":
		return readBinaryVersion(proxyBin)
	case "zid-geolocation":
		return readBinaryVersion(geolocationBin)
	case "zid-logs":
		if xmlv := readPackageXMLVersion("/usr/local/pkg/zid-logs.xml"); xmlv != "" {
			return xmlv
		}
		if v := readVersionFile("/usr/local/share/pfSense-pkg-zid-logs/VERSION"); v != "" {
			return v
		}
		// config.xml pode conter version "dev" (ex.: "zid-logs version dev") dependendo de como o pacote foi registrado.
		// Para exibir/comparar updates, precisamos de uma versao numerica.
		if v := readConfigXMLPackageVersion("zid-logs"); v != "" {
			if nv := extractNumericVersion(v); nv != "" {
				return nv
			}
		}
		return readBinaryVersion(logsBin)
	case "zid-access":
		if v := readConfigXMLPackageVersion("zid-access"); v != "" {
			return v
		}
		if v := readVersionFile("/usr/local/share/pfSense-pkg-zid-access/VERSION"); v != "" {
			return v
		}
		return ""
	case "zid-orchestrator":
		if v := readConfigXMLPackageVersion("zid-orchestration"); v != "" {
			return v
		}
		if v := readVersionFile("/usr/local/share/pfSense-pkg-zid-orchestration/VERSION"); v != "" {
			return v
		}
		return readBinaryVersion(orchestratorBin)
	default:
		return ""
	}
}

func VersionRemote(key string) string {
	pkg, err := Get(key)
	if err != nil {
		return ""
	}
	ver, err := s3.FetchVersion(pkg.VersionURL)
	if err != nil {
		return ""
	}
	return ver
}

func UpdateAvailable(key string) bool {
	local := VersionLocal(key)
	remote := VersionRemote(key)
	return UpdateAvailableWith(local, remote)
}

func UpdateAvailableWith(local, remote string) bool {
	if local == "" || remote == "" {
		return false
	}
	return compareVersion(remote, local) > 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func pgrepRunning(pattern string) bool {
	cmd := exec.Command("/usr/bin/pgrep", "-f", pattern)
	return cmd.Run() == nil
}

func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func readBinaryVersion(bin string) string {
	if !fileExists(bin) {
		return ""
	}
	for _, arg := range []string{"-version", "--version"} {
		out, err := exec.Command(bin, arg).CombinedOutput()
		if err == nil {
			return parseVersion(string(out))
		}
	}
	return ""
}

func parseVersion(output string) string {
	re := regexp.MustCompile(`\b(\d+(?:\.\d+)+)\b`)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		return match[1]
	}
	return strings.TrimSpace(strings.Split(output, "\n")[0])
}

func extractNumericVersion(output string) string {
	re := regexp.MustCompile(`(\d+(?:\.\d+)+)`)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func compareVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	max := len(as)
	if len(bs) > max {
		max = len(bs)
	}
	for i := 0; i < max; i++ {
		ai := 0
		bi := 0
		if i < len(as) {
			fmt.Sscanf(as[i], "%d", &ai)
		}
		if i < len(bs) {
			fmt.Sscanf(bs[i], "%d", &bi)
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func readPackageXMLVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`<version>([^<]+)</version>`)
	match := re.FindStringSubmatch(string(data))
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func readVersionFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readConfigXMLPackageVersion(pkgName string) string {
	data, err := os.ReadFile(configXMLPath)
	if err != nil {
		return ""
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var stack []string
	var inPackage bool
	var currentName string
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			if matchesPath(stack, []string{"installedpackages", "package"}) {
				inPackage = true
				currentName = ""
			}
		case xml.EndElement:
			if matchesPath(stack, []string{"installedpackages", "package"}) {
				inPackage = false
				currentName = ""
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if inPackage && matchesPath(stack, []string{"installedpackages", "package", "name"}) {
				currentName = strings.TrimSpace(string(t))
			}
			if inPackage && currentName == pkgName && matchesPath(stack, []string{"installedpackages", "package", "version"}) {
				val := strings.TrimSpace(string(t))
				if val != "" {
					return val
				}
			}
		}
	}
}

func isOn(val string) bool {
	val = strings.ToLower(strings.TrimSpace(val))
	return val == "on" || val == "true" || val == "1" || val == "yes"
}

func logEnable(key, source, val string, ok bool) {
	if !enableDebug {
		return
	}
	msg := "enable read: key=" + key + " source=" + source + " ok=" + boolString(ok) + " value=" + strings.TrimSpace(val)
	enableLogger.Info(msg)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func envTrue(name string) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	return val == "1" || val == "true" || val == "yes"
}

func cacheEnabled(key string, value bool) bool {
	enabledCacheMu.Lock()
	defer enabledCacheMu.Unlock()
	enabledCache[key] = enabledCacheEntry{value: value, timestamp: time.Now().UTC()}
	return value
}

func cachedEnabled(key string) (bool, bool) {
	enabledCacheMu.Lock()
	defer enabledCacheMu.Unlock()
	entry, ok := enabledCache[key]
	if !ok {
		return false, false
	}
	if time.Since(entry.timestamp) > enabledCacheTTL {
		delete(enabledCache, key)
		return false, false
	}
	return entry.value, true
}

func readEnableViaPHP(key string) (bool, bool) {
	php := phpBin()
	if php == "" {
		return false, false
	}
	var expr string
	switch key {
	case "zid-proxy":
		expr = `$cfg=$config["installedpackages"]["zidproxy"]["config"][0] ?? []; $val=$cfg["enable"] ?? ""; echo ($val === "on" || $val === "true" || $val === "1" || $val === true || $val === 1) ? "1" : "0";`
	case "zid-threatd":
		expr = `$cfg=$config["installedpackages"]["zidproxy"]["config"][0] ?? []; $val=$cfg["threat_enable"] ?? ""; echo ($val === "on" || $val === "true" || $val === "1" || $val === true || $val === 1) ? "1" : "0";`
	case "zid-geolocation":
		expr = `$cfg=$config["installedpackages"]["zidgeolocation"]["config"][0] ?? []; $val=$cfg["enable"] ?? ""; echo ($val === "on" || $val === "true" || $val === "1" || $val === true || $val === 1) ? "1" : "0";`
	case "zid-access":
		expr = `$raw=null;
if (isset($config["installedpackages"]["zidaccess"]["config"])) { $raw=$config["installedpackages"]["zidaccess"]["config"]; }
elseif (isset($config["installedpackages"]["zid-access"]["config"])) { $raw=$config["installedpackages"]["zid-access"]["config"]; }
elseif (isset($config["installedpackages"]["zid_access"]["config"])) { $raw=$config["installedpackages"]["zid_access"]["config"]; }
$val="";
// Formato quebrado (legado): lista escalar sem chaves. Primeiro item costuma ser o enable.
if (is_array($raw) && isset($raw[0]) && !is_array($raw[0])) {
  $val=$raw[0];
} elseif (!is_array($raw) && $raw !== null) {
  $val=$raw;
} else {
  $item=$raw;
  if (is_array($raw) && isset($raw[0]) && is_array($raw[0])) { $item=$raw[0]; }
  elseif (is_array($raw) && isset($raw["item"])) {
    $i=$raw["item"];
    if (is_array($i) && isset($i[0]) && is_array($i[0])) { $item=$i[0]; }
    elseif (is_array($i)) { $item=$i; }
  }
  if (is_array($item)) {
    if (array_key_exists("enable", $item)) { $val=$item["enable"]; }
    elseif (array_key_exists("enabled", $item)) { $val=$item["enabled"]; }
  }
}
echo ($val === "on" || $val === "true" || $val === "1" || $val === "yes" || $val === true || $val === 1) ? "1" : "0";`
	default:
		return false, false
	}
	cmd := exec.Command(php, "-r", `require_once("/etc/inc/config.inc"); `+expr)
	out, err := cmd.Output()
	if err != nil {
		return false, false
	}
	val := strings.TrimSpace(string(out))
	if val == "1" {
		return true, true
	}
	if val == "0" {
		return false, true
	}
	return false, false
}

func phpBin() string {
	if fileExists("/usr/local/bin/php") {
		return "/usr/local/bin/php"
	}
	if fileExists("/usr/local/bin/php-cgi") {
		return "/usr/local/bin/php-cgi"
	}
	return ""
}

func readValueOrEmpty(path []string) string {
	if val, ok := readConfigXMLValueRetry(path, 3); ok {
		return strings.TrimSpace(val)
	}
	return ""
}

func readValueLooseOrEmpty(path []string) string {
	if val, ok := readConfigXMLValueLooseRetry(path, 3); ok {
		return strings.TrimSpace(val)
	}
	return ""
}

func readRCConfBool(path, key string) (bool, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	lines := strings.Split(string(data), "\n")
	prefix := key + "="
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			val = strings.Trim(val, "\"'")
			return isOn(val), true
		}
	}
	return false, false
}

func startAppID() error {
	if !fileExists(appidBin) {
		return errors.New("appid binary not found")
	}
	_ = stopAppID()
	_ = os.MkdirAll("/usr/local/etc/zid-proxy", 0755)
	rules := "/usr/local/etc/zid-proxy/appid_rules.txt"
	if !fileExists(rules) {
		_ = os.WriteFile(rules, []byte(""), 0644)
	}
	_ = os.Remove("/var/run/zid-appid.sock")
	args := []string{
		"-f",
		"-p", "/var/run/zid-appid.pid",
		appidBin,
		"-socket", "/var/run/zid-appid.sock",
		"-rules", rules,
		"-pid", "/var/run/zid-appid.pid",
	}
	cmd := exec.Command("/usr/sbin/daemon", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stopAppID() error {
	pids, _ := exec.Command("/usr/bin/pgrep", "-f", "^/usr/local/sbin/zid-appid").Output()
	if len(pids) == 0 {
		return nil
	}
	lines := strings.Fields(string(pids))
	for _, pid := range lines {
		_ = exec.Command("/bin/kill", pid).Run()
	}
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if !pgrepRunning("^/usr/local/sbin/zid-appid") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if pgrepRunning("^/usr/local/sbin/zid-appid") {
		for _, pid := range lines {
			_ = exec.Command("/bin/kill", "-9", pid).Run()
		}
	}
	_ = os.Remove("/var/run/zid-appid.pid")
	_ = os.Remove("/var/run/zid-appid.sock")
	return nil
}
