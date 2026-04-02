package main

import "time"

func (w *winboatApp) loadInitialState() {
	cfg, err := loadStoredConfig()
	if err != nil {
		w.reportError("Loading saved settings", err)
	}

	w.usernameEntry.SetText(cfg.Username)
	w.passwordEntry.SetText(cfg.Password)
	w.scaleSelect.SetSelected(normalizeScale(cfg.Scale))
	w.preferredMonitors = append([]int(nil), cfg.MonitorIDs...)
	w.loadStartupSettings()

	w.reloadMonitors(false)
	w.appendLog("Loaded saved credentials and monitor preferences from %s.", configDir())
	w.refreshStatusAsync(false)
}

func (w *winboatApp) startStatusPolling() {
	go func() {
		ticker := time.NewTicker(statusPollEvery)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.refreshStatusAsync(false)
			case <-w.done:
				return
			}
		}
	}()
}

func (w *winboatApp) signalDone() {
	w.quitOnce.Do(func() {
		close(w.done)
	})
}

func (w *winboatApp) notifyHideToTray() {}

func (w *winboatApp) loadStartupSettings() {
	settings, err := currentAutostartSettings(w.app)
	if err != nil {
		w.reportError("Loading startup settings", err)
		return
	}

	w.launchAtLoginCheck.SetChecked(settings.Enabled)
}
