package status

import (
	"time"

	"zid-packages/internal/autoupdate"
	"zid-packages/internal/licensing"
	"zid-packages/internal/packages"
)

type PackageStatus struct {
	Key                     string `json:"key"`
	Installed               bool   `json:"installed"`
	Enabled                 bool   `json:"enabled"`
	Licensed                bool   `json:"licensed"`
	ServiceRunning          bool   `json:"service_running"`
	VersionLocal            string `json:"version_installed"`
	VersionRemote           string `json:"version_remote"`
	UpdateAvailable         bool   `json:"update_available"`
	AutoUpdateAgeDays       int    `json:"auto_update_age_days"`
	AutoUpdateThresholdDays int    `json:"auto_update_threshold_days"`
	AutoUpdateDue           bool   `json:"auto_update_due"`
	AutoUpdateDueAt         int64  `json:"auto_update_due_at"`
}

type ServiceStatus struct {
	Key         string `json:"key"`
	Running     bool   `json:"running"`
	Enabled     bool   `json:"enabled"`
	Licensed    bool   `json:"licensed"`
	DisplayName string `json:"name"`
}

type LicensingStatus struct {
	LastAttempt int64  `json:"last_attempt"`
	LastSuccess int64  `json:"last_success"`
	ValidUntil  int64  `json:"valid_until"`
	Mode        string `json:"mode"`
	Reason      string `json:"reason"`
}

type Status struct {
	Packages  []PackageStatus `json:"packages"`
	Services  []ServiceStatus `json:"services"`
	Licensing LicensingStatus `json:"licensing"`
}

func BuildStatus() Status {
	pkgs := packages.All()
	out := make([]PackageStatus, 0, len(pkgs))

	st, err := licensing.LoadState()
	now := time.Now()
	autoState, _ := autoupdate.Load()
	autoChanged := false
	mode := licensing.ModeNeverOK
	validUntil := time.Time{}
	reason := "never_ok"
	if err == nil {
		mode, validUntil = licensing.Evaluate(st, now)
		reason = modeReason(mode)
	} else {
		reason = err.Error()
	}
	licenseOK := mode == licensing.ModeOK || mode == licensing.ModeOfflineGrace

	for _, pkg := range pkgs {
		licensed := false
		if st.Licensed != nil && licenseOK {
			licensed = st.Licensed[pkg.Key]
		}
		enabled, _ := packages.Enabled(pkg.Key)
		running, _ := packages.ServiceRunning(pkg.Key)
		localVersion := packages.VersionLocal(pkg.Key)
		remoteVersion := packages.VersionRemote(pkg.Key)
		updateAvailable := packages.UpdateAvailableWith(localVersion, remoteVersion)
		entry, changed := autoupdate.Update(&autoState, pkg.Key, updateAvailable, remoteVersion, now)
		if changed {
			autoChanged = true
		}
		autoAge := autoupdate.AgeDays(entry, now)
		autoDue := autoupdate.Due(entry, now)
		autoDueAt := autoupdate.DueAt(entry, autoupdate.ThresholdDays(), time.Local)
		out = append(out, PackageStatus{
			Key:                     pkg.Key,
			Installed:               packages.Installed(pkg.Key),
			Enabled:                 enabled,
			Licensed:                licensed,
			ServiceRunning:          running,
			VersionLocal:            localVersion,
			VersionRemote:           remoteVersion,
			UpdateAvailable:         updateAvailable,
			AutoUpdateAgeDays:       autoAge,
			AutoUpdateThresholdDays: autoupdate.ThresholdDays(),
			AutoUpdateDue:           autoDue,
			AutoUpdateDueAt:         unixOrZero(autoDueAt),
		})
	}

	if autoChanged {
		_ = autoupdate.Save(autoState)
	}

	services := buildServicesStatus(st.Licensed, licenseOK)

	return Status{
		Packages: out,
		Services: services,
		Licensing: LicensingStatus{
			LastAttempt: st.LastAttempt.Unix(),
			LastSuccess: st.LastSuccess.Unix(),
			ValidUntil:  unixOrZero(validUntil),
			Mode:        mode,
			Reason:      reason,
		},
	}
}

func buildServicesStatus(licensed map[string]bool, licenseOK bool) []ServiceStatus {
	services := []ServiceStatus{}

	services = append(services, serviceStatus("zid-packages", "zid-packages", "", licensed, licenseOK))
	services = append(services, serviceStatus("zid-proxy", "zid-proxy", "", licensed, licenseOK))
	services = append(services, serviceStatus("zid-appid", "zid-proxy", "", licensed, licenseOK))
	services = append(services, serviceStatus("zid-threatd", "zid-proxy", "zid-threatd", licensed, licenseOK))
	services = append(services, serviceStatus("zid-geolocation", "zid-geolocation", "", licensed, licenseOK))
	services = append(services, serviceStatus("zid-logs", "zid-logs", "", licensed, licenseOK))

	return services
}

func serviceStatus(key, packageKey, enableKey string, licensed map[string]bool, licenseOK bool) ServiceStatus {
	if enableKey == "" {
		enableKey = packageKey
	}
	enabled, _ := packages.Enabled(enableKey)
	running, _ := packages.ServiceRunning(key)
	licensedOK := licenseOK && licensed[packageKey]
	return ServiceStatus{
		Key:         key,
		DisplayName: key,
		Running:     running,
		Enabled:     enabled,
		Licensed:    licensedOK,
	}
}

func unixOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func modeReason(mode string) string {
	switch mode {
	case licensing.ModeOK:
		return "ok"
	case licensing.ModeOfflineGrace:
		return "offline_grace"
	case licensing.ModeExpired:
		return "expired"
	default:
		return "never_ok"
	}
}
