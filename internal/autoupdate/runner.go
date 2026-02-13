package autoupdate

import (
	"time"

	"zid-packages/internal/logx"
	"zid-packages/internal/packages"
)

func RunOnce(logger *logx.Logger, now time.Time) {
	st, _ := Load()
	changed := false
	// Atualiza o próprio zid-packages por último. Mesmo com update seguro, isso
	// reduz o risco de interromper a rodada quando algum ambiente ainda reinicia
	// o daemon durante a instalação.
	all := packages.All()
	ordered := make([]packages.Package, 0, len(all))
	var self *packages.Package
	for _, pkg := range all {
		if pkg.Key == "zid-packages" {
			p := pkg
			self = &p
			continue
		}
		ordered = append(ordered, pkg)
	}
	if self != nil {
		ordered = append(ordered, *self)
	}

	for _, pkg := range ordered {
		if !packages.Installed(pkg.Key) {
			if Clear(&st, pkg.Key) {
				changed = true
			}
			continue
		}
		remoteVersion := packages.VersionRemote(pkg.Key)
		localVersion := packages.VersionLocal(pkg.Key)
		updateAvailable := packages.UpdateAvailableWith(localVersion, remoteVersion)
		entry, updated := Update(&st, pkg.Key, updateAvailable, remoteVersion, now)
		if updated {
			changed = true
		}
		if !updateAvailable {
			continue
		}
		if !Due(entry, now) {
			continue
		}
		logger.Info("auto-update start: " + pkg.Key)
		if err := packages.Update(logger, pkg.Key); err != nil {
			logger.Error("auto-update failed: " + pkg.Key + " err=" + err.Error())
			continue
		}
		if Clear(&st, pkg.Key) {
			changed = true
		}
		logger.Info("auto-update done: " + pkg.Key)
	}
	MarkRun(&st, now)
	if changed {
		_ = Save(st)
	} else {
		_ = Save(st)
	}
}
