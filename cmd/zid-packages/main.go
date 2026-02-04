package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"zid-packages/internal/autoupdate"
	"zid-packages/internal/licensing"
	"zid-packages/internal/logx"
	"zid-packages/internal/packages"
	"zid-packages/internal/status"
	"zid-packages/internal/watchdog"
)

var version = "dev"

func main() {
	logger := logx.New("/var/log/zid-packages.log")
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	if cmd == "-version" || cmd == "version" {
		fmt.Println("zid-packages", version)
		return
	}
	switch cmd {
	case "status":
		handleStatus(os.Args[2:])
	case "watchdog":
		handleWatchdog(logger, os.Args[2:])
	case "license":
		handleLicense(logger, os.Args[2:])
	case "package":
		handlePackage(logger, os.Args[2:])
	case "daemon":
		handleDaemon(logger, os.Args[2:])
	case "auto-update":
		handleAutoUpdate(logger, os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "zid-packages <command> [args]")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  status [--json]")
	fmt.Fprintln(os.Stderr, "  watchdog --once")
	fmt.Fprintln(os.Stderr, "  license sync")
	fmt.Fprintln(os.Stderr, "  package install <pkg>")
	fmt.Fprintln(os.Stderr, "  package update <pkg>")
	fmt.Fprintln(os.Stderr, "  auto-update --once")
	fmt.Fprintln(os.Stderr, "  daemon")
}

func handleStatus(args []string) {
	jsonFlag := len(args) == 1 && args[0] == "--json"
	st := status.BuildStatus()
	if jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(st); err != nil {
			fmt.Fprintln(os.Stderr, "failed to encode status:", err)
			os.Exit(1)
		}
		return
	}
	fmt.Println("Use --json")
}

func handleWatchdog(logger *logx.Logger, args []string) {
	if len(args) != 1 || args[0] != "--once" {
		usage()
		os.Exit(2)
	}
	if err := watchdog.RunOnce(logger); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func handleLicense(logger *logx.Logger, args []string) {
	if len(args) != 1 || args[0] != "sync" {
		usage()
		os.Exit(2)
	}
	if err := licensing.Sync(logger); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func handlePackage(logger *logx.Logger, args []string) {
	if len(args) != 2 {
		usage()
		os.Exit(2)
	}
	action := args[0]
	key := strings.TrimSpace(args[1])
	if key == "" {
		fmt.Fprintln(os.Stderr, "package key required")
		os.Exit(2)
	}
	if err := packages.ValidateKey(key); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	var err error
	switch action {
	case "install":
		err = packages.Install(logger, key)
	case "update":
		err = packages.Update(logger, key)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func handleDaemon(logger *logx.Logger, args []string) {
	if len(args) != 0 {
		usage()
		os.Exit(2)
	}
	if err := watchdog.RunDaemon(logger, 0); err != nil {
		if !errors.Is(err, watchdog.ErrDaemonStopped) {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func handleAutoUpdate(logger *logx.Logger, args []string) {
	if len(args) != 1 || args[0] != "--once" {
		usage()
		os.Exit(2)
	}
	now := time.Now().UTC()
	autoupdate.RunOnce(logger, now)
}
