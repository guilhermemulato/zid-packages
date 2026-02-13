package packages

import (
	"os"
	"strings"
)

const restartPendingPath = "/var/db/zid-packages/restart-pending"

func RestartPendingInfo() (bool, string) {
	data, err := os.ReadFile(restartPendingPath)
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(data))
}

func ClearRestartPending() {
	_ = os.Remove(restartPendingPath)
}
