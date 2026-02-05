package watchdog

import (
	"errors"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"zid-packages/internal/autoupdate"
	"zid-packages/internal/ipc"
	"zid-packages/internal/licensing"
	"zid-packages/internal/logx"
	"zid-packages/internal/packages"
)

var ErrDaemonStopped = errors.New("daemon stopped")

const (
	watchdogInterval = 1 * time.Minute
	licenseInterval  = 2 * time.Hour
)

type service struct {
	Key         string
	PackageKey  string
	EnableKey   string
	DisplayName string
}

var services = []service{
	{Key: "zid-proxy", PackageKey: "zid-proxy", DisplayName: "zid-proxy"},
	{Key: "zid-appid", PackageKey: "zid-proxy", DisplayName: "zid-appid"},
	{Key: "zid-threatd", PackageKey: "zid-proxy", EnableKey: "zid-threatd", DisplayName: "zid-threatd"},
	{Key: "zid-geolocation", PackageKey: "zid-geolocation", DisplayName: "zid-geolocation"},
	{Key: "zid-logs", PackageKey: "zid-logs", DisplayName: "zid-logs"},
	{Key: "zid-access", PackageKey: "zid-access", DisplayName: "zid-access"},
}

func RunOnce(logger *logx.Logger) error {
	now := time.Now().UTC()
	st, err := licensing.LoadState()
	if err != nil {
		logger.Error("falha ao carregar state: " + err.Error())
	}
	if st.LastAttempt.IsZero() || now.Sub(st.LastAttempt) >= licenseInterval {
		if err := licensing.Sync(logger); err != nil {
			logger.Error("licensing sync falhou: " + err.Error())
		}
		st, _ = licensing.LoadState()
	}

	mode, _ := licensing.Evaluate(st, now)
	licenseOK := mode == licensing.ModeOK || mode == licensing.ModeOfflineGrace

	for _, svc := range services {
		if !packages.Installed(svc.PackageKey) {
			continue
		}
		enableKey := svc.PackageKey
		if svc.EnableKey != "" {
			enableKey = svc.EnableKey
		}
		enabled, _ := packages.Enabled(enableKey)
		licensed := licenseOK && st.Licensed[svc.PackageKey]
		shouldRun := enabled && licensed

		running, _ := packages.ServiceRunning(svc.Key)
		if shouldRun && !running {
			logger.Info("watchdog start: " + svc.DisplayName + watchdogReason(enabled, licensed, mode))
			_ = packages.StartService(svc.Key)
		}
		if !shouldRun && running {
			logger.Info("watchdog stop: " + svc.DisplayName + watchdogReason(enabled, licensed, mode))
			if !enabled {
				logger.Info("watchdog enable snapshot: " + svc.DisplayName + " " + formatEnableSnapshot(enableKey))
			}
			_ = packages.StopService(svc.Key)
		}
	}
	return nil
}

func RunDaemon(logger *logx.Logger, interval time.Duration) error {
	if interval <= 0 {
		interval = watchdogInterval
	}

	ipcServer := ipc.NewServer(logger)
	if err := ipcServer.Start(); err != nil {
		return err
	}
	defer ipcServer.Stop()

	watchdogTicker := time.NewTicker(interval)
	licenseTicker := time.NewTicker(licenseInterval)
	defer watchdogTicker.Stop()
	defer licenseTicker.Stop()

	logger.Info("daemon start")
	_ = RunOnce(logger)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-watchdogTicker.C:
			_ = RunOnce(logger)
			nowLocal := time.Now()
			autoState, _ := autoupdate.Load()
			if autoupdate.ShouldRunNow(autoState, nowLocal, autoupdate.ScheduleHour, autoupdate.ScheduleMinute) {
				autoupdate.RunOnce(logger, nowLocal)
			}
		case <-licenseTicker.C:
			_ = licensing.Sync(logger)
		case <-sigs:
			logger.Info("daemon stop")
			return ErrDaemonStopped
		}
	}
}

func watchdogReason(enabled bool, licensed bool, mode string) string {
	return " (enabled=" + boolLabel(enabled) + " licensed=" + boolLabel(licensed) + " mode=" + mode + ")"
}

func boolLabel(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func formatEnableSnapshot(packageKey string) string {
	snapshot := packages.EnableSnapshot(packageKey)
	if len(snapshot) == 0 {
		return "(no snapshot)"
	}
	parts := make([]string, 0, len(snapshot))
	for key, val := range snapshot {
		if val == "" {
			val = "-"
		}
		parts = append(parts, key+"="+val)
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}
