package packages

import (
	"errors"
	"fmt"
	"strings"

	"zid-packages/internal/logx"
)

type Package struct {
	Key               string
	Name              string
	BundleURL         string
	VersionURL        string
	UpdateCommand     string
	InstallScriptGlob string
}

var all = []Package{
	{
		Key:               "zid-packages",
		Name:              "ZID Packages",
		BundleURL:         "https://s3.soulsolucoes.com.br/soul/portal/zid-packages-latest.tar.gz",
		VersionURL:        "https://s3.soulsolucoes.com.br/soul/portal/zid-packages-latest.version",
		UpdateCommand:     "/usr/local/sbin/zid-packages-update",
		InstallScriptGlob: "*/scripts/install.sh",
	},
	{
		Key:               "zid-proxy",
		Name:              "ZID Proxy",
		BundleURL:         "https://s3.soulsolucoes.com.br/soul/portal/zid-proxy-pfsense-latest.tar.gz",
		VersionURL:        "https://s3.soulsolucoes.com.br/soul/portal/zid-proxy-pfsense-latest.version",
		UpdateCommand:     "/usr/local/sbin/zid-proxy-update",
		InstallScriptGlob: "*/pkg-zid-proxy/install.sh",
	},
	{
		Key:               "zid-geolocation",
		Name:              "ZID Geolocation",
		BundleURL:         "https://s3.soulsolucoes.com.br/soul/portal/zid-geolocation-latest.tar.gz",
		VersionURL:        "https://s3.soulsolucoes.com.br/soul/portal/zid-geolocation-latest.version",
		UpdateCommand:     "/usr/local/sbin/zid-geolocation-update",
		InstallScriptGlob: "*/scripts/install.sh",
	},
	{
		Key:               "zid-logs",
		Name:              "ZID Logs",
		BundleURL:         "https://s3.soulsolucoes.com.br/soul/portal/zid-logs-latest.tar.gz",
		VersionURL:        "https://s3.soulsolucoes.com.br/soul/portal/zid-logs-latest.version",
		UpdateCommand:     "/usr/local/sbin/zid-logs-update",
		InstallScriptGlob: "*/pkg-zid-logs/install.sh",
	},
}

func All() []Package {
	return append([]Package(nil), all...)
}

func Get(key string) (Package, error) {
	key = strings.TrimSpace(key)
	for _, pkg := range all {
		if pkg.Key == key {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("unknown package: %s", key)
}

func ValidateKey(key string) error {
	_, err := Get(key)
	return err
}

func Install(logger *logx.Logger, key string) error {
	pkg, err := Get(key)
	if err != nil {
		return err
	}
	logger.Info("install requested: " + pkg.Key)
	return installBundle(pkg)
}

func Update(logger *logx.Logger, key string) error {
	pkg, err := Get(key)
	if err != nil {
		return err
	}
	logger.Info("update requested: " + pkg.Key)
	if pkg.UpdateCommand == "" {
		return errors.New("update command not defined")
	}
	return runUpdate(pkg.UpdateCommand)
}
