package watchdog

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	DisplayName string
}

var services = []service{
	{Key: "zid-proxy", PackageKey: "zid-proxy", DisplayName: "zid-proxy"},
	{Key: "zid-appid", PackageKey: "zid-proxy", DisplayName: "zid-appid"},
	{Key: "zid-geolocation", PackageKey: "zid-geolocation", DisplayName: "zid-geolocation"},
	{Key: "zid-logs", PackageKey: "zid-logs", DisplayName: "zid-logs"},
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
		enabled, _ := packages.Enabled(svc.PackageKey)
		licensed := licenseOK && st.Licensed[svc.PackageKey]
		shouldRun := enabled && licensed

		running, _ := packages.ServiceRunning(svc.Key)
		if shouldRun && !running {
			logger.Info("watchdog start: " + svc.DisplayName)
			_ = packages.StartService(svc.Key)
		}
		if !shouldRun && running {
			logger.Info("watchdog stop: " + svc.DisplayName)
			_ = packages.StopService(svc.Key)
		}
	}
	return nil
}

func RunDaemon(logger *logx.Logger, interval time.Duration) error {
	if interval <= 0 {
		interval = watchdogInterval
	}

	ipcServer := ipc.NewServer()
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
		case <-licenseTicker.C:
			_ = licensing.Sync(logger)
		case <-sigs:
			logger.Info("daemon stop")
			return ErrDaemonStopped
		}
	}
}
