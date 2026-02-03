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
	proxyBin       = "/usr/local/sbin/zid-proxy"
	appidBin       = "/usr/local/sbin/zid-appid"
	geolocationBin = "/usr/local/sbin/zid-geolocation"
	logsBin        = "/usr/local/sbin/zid-logs"
	packagesBin    = "/usr/local/sbin/zid-packages"
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
	case "zid-geolocation":
		out["config:installedpackages/zidgeolocation/config/enable"] = readValueOrEmpty([]string{"installedpackages", "zidgeolocation", "config", "enable"})
		out["config:zidgeolocation/config/enable"] = readValueOrEmpty([]string{"zidgeolocation", "config", "enable"})
		out["config-loose:installedpackages/zidgeolocation/config/enable"] = readValueLooseOrEmpty([]string{"installedpackages", "zidgeolocation", "config", "enable"})
		out["config-loose:zidgeolocation/config/enable"] = readValueLooseOrEmpty([]string{"zidgeolocation", "config", "enable"})
		if b, ok := readJSONBool("/usr/local/etc/zid-geolocation/config.json", "enable"); ok {
			out["config-json:/usr/local/etc/zid-geolocation/config.json"] = boolString(b)
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
	case "zid-appid":
		return pgrepRunning("^/usr/local/sbin/zid-appid"), nil
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
		return run("/usr/local/etc/rc.d/zid_geolocation", "onestart")
	case "zid-logs":
		return run("/usr/local/etc/rc.d/zid_logs", "onestart")
	case "zid-appid":
		return startAppID()
	default:
		return errors.New("unknown service")
	}
}

func StopService(key string) error {
	switch key {
	case "zid-packages":
		return run("/usr/local/etc/rc.d/zid_packages", "onestop")
	case "zid-proxy":
		return run("/usr/local/etc/rc.d/zid-proxy.sh", "stop")
	case "zid-geolocation":
		return run("/usr/local/etc/rc.d/zid_geolocation", "onestop")
	case "zid-logs":
		return run("/usr/local/etc/rc.d/zid_logs", "onestop")
	case "zid-appid":
		return stopAppID()
	default:
		return errors.New("unknown service")
	}
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
		if v := readConfigXMLPackageVersion("zid-logs"); v != "" {
			return v
		}
		if xmlv := readPackageXMLVersion("/usr/local/pkg/zid-logs.xml"); xmlv != "" {
			return xmlv
		}
		return readBinaryVersion(logsBin)
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
	out, err := exec.Command(bin, "-version").CombinedOutput()
	if err != nil {
		return ""
	}
	return parseVersion(string(out))
}

func parseVersion(output string) string {
	re := regexp.MustCompile(`\b(\d+\.\d+(?:\.\d+){0,3})\b`)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		return match[1]
	}
	return strings.TrimSpace(strings.Split(output, "\n")[0])
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
	case "zid-geolocation":
		expr = `$cfg=$config["installedpackages"]["zidgeolocation"]["config"][0] ?? []; $val=$cfg["enable"] ?? ""; echo ($val === "on" || $val === "true" || $val === "1" || $val === true || $val === 1) ? "1" : "0";`
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
