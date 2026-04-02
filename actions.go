package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
)

func (w *winboatApp) connect() {
	w.connectFlow("Connect", false)
}

func (w *winboatApp) restartAndConnect() {
	w.connectFlow("Restart and Connect", true)
}

func (w *winboatApp) connectFlow(action string, restartFirst bool) {
	cfg, err := w.currentConfigFromUI()
	if err != nil {
		w.reportError(action, err)
		return
	}

	if !w.beginBusy(action) {
		return
	}

	go func() {
		defer w.endBusy("Ready")

		if err := saveStoredConfig(storedConfig{Username: cfg.Username, Password: cfg.Password, MonitorIDs: cfg.MonitorIDs}); err != nil {
			w.reportError(action, fmt.Errorf("save settings: %w", err))
			return
		}

		if restartFirst {
			w.setLastAction("Restarting VM...")
			w.appendLog("Restarting %s container before launching RDP.", containerName)
			if _, err := runCommand(context.Background(), 30*time.Second, "docker", "restart", containerName); err != nil {
				w.reportError(action, fmt.Errorf("restart container: %w", err))
				return
			}
			w.appendLog("WinBoat container restarted.")
		}

		w.setLastAction("Waiting for RDP port...")
		port, err := w.waitForPort(portWaitTimeout)
		if err != nil {
			w.reportError(action, err)
			return
		}

		w.setLastAction("Launching FreeRDP...")
		if err := w.launchRDP(cfg, port); err != nil {
			w.reportError(action, err)
			return
		}

		w.refreshStatusAsync(false)
	}()
}

func (w *winboatApp) toggleVM() {
	w.mu.Lock()
	status := w.containerStatus
	w.mu.Unlock()

	action := "Start VM"
	command := "start"
	successMessage := "WinBoat container started."
	if isRunningStatus(status) {
		action = "Stop VM"
		command = "stop"
		successMessage = "WinBoat container stopped."
	}

	if !w.beginBusy(action) {
		return
	}

	go func() {
		defer w.endBusy("Ready")

		w.appendLog("Running docker %s on %s.", command, containerName)
		if _, err := runCommand(context.Background(), 30*time.Second, "docker", command, containerName); err != nil {
			w.reportError(action, fmt.Errorf("%s container: %w", command, err))
			return
		}

		w.appendLog("%s", successMessage)
		w.refreshStatusAsync(false)
	}()
}

func (w *winboatApp) saveSettings() {
	cfg, err := w.currentConfigFromUI()
	if err != nil {
		w.reportError("Save Settings", err)
		return
	}

	if !w.beginBusy("Save Settings") {
		return
	}

	go func() {
		defer w.endBusy("Ready")

		if err := saveStoredConfig(storedConfig{Username: cfg.Username, Password: cfg.Password, MonitorIDs: cfg.MonitorIDs}); err != nil {
			w.reportError("Save Settings", err)
			return
		}

		if err := setAutostart(w.app, autostartSettings{
			Enabled: w.launchAtLoginCheck.Checked,
		}); err != nil {
			w.reportError("Save Settings", err)
			return
		}

		w.appendLog("Saved credentials and monitor selection to %s.", configDir())
		if w.launchAtLoginCheck.Checked {
			w.appendLog("Autostart enabled.")
		} else {
			w.appendLog("Autostart disabled.")
		}
		w.refreshSelectedLabel(cfg.MonitorIDs)
	}()
}

func (w *winboatApp) clearStoredCredentials() {
	if !w.beginBusy("Clear Stored Credentials") {
		return
	}

	go func() {
		defer w.endBusy("Ready")

		if err := clearCredentials(); err != nil {
			w.reportError("Clear Stored Credentials", err)
			return
		}

		fyne.Do(func() {
			w.usernameEntry.SetText("")
			w.passwordEntry.SetText("")
		})
		w.appendLog("Removed stored credentials from %s.", configFilePath(credentialsFile))
	}()
}

func (w *winboatApp) reloadMonitors(showResult bool) {
	go func() {
		currentSelection := w.selectedMonitorIDsFromUI()
		options, backend, err := detectMonitors()
		if err != nil {
			w.reportError("Refresh Monitors", err)
			return
		}

		fyne.Do(func() {
			w.monitorOptions = options
			w.labelToMonitor = make(map[string]int, len(options))
			labels := make([]string, 0, len(options))
			for _, option := range options {
				labels = append(labels, option.Label)
				w.labelToMonitor[option.Label] = option.ID
			}

			w.monitorChecks.Options = labels
			w.monitorChecks.Refresh()

			selected := currentSelection
			if len(selected) == 0 {
				selected = append([]int(nil), w.preferredMonitors...)
			}
			if len(selected) == 0 {
				selected = allMonitorIDs(options)
			}
			w.preferredMonitors = nil
			w.applySelectedMonitors(selected)
			w.monitorHint.SetText(fmt.Sprintf("Pick the monitors to use for the RDP session. %d detected.", len(options)))
		})

		if showResult {
			w.appendLog("Reloaded %d monitor option(s) using %s.", len(options), backend.DisplayName)
		}
	}()
}
