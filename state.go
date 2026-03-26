package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (w *winboatApp) currentConfigFromUI() (runtimeConfig, error) {
	username := w.usernameEntry.Text
	password := w.passwordEntry.Text
	monitorIDs := w.selectedMonitorIDsFromUI()

	if strings.TrimSpace(username) == "" {
		return runtimeConfig{}, errors.New("username is required")
	}

	if password == "" {
		return runtimeConfig{}, errors.New("password is required")
	}

	if len(monitorIDs) == 0 {
		return runtimeConfig{}, errors.New("select at least one monitor")
	}

	return runtimeConfig{
		Username:   username,
		Password:   password,
		MonitorIDs: monitorIDs,
	}, nil
}

func (w *winboatApp) selectedMonitorIDsFromUI() []int {
	w.mu.Lock()
	options := append([]monitorOption(nil), w.monitorOptions...)
	w.mu.Unlock()

	selected := make(map[string]bool, len(w.monitorChecks.Selected))
	for _, label := range w.monitorChecks.Selected {
		selected[label] = true
	}

	ids := make([]int, 0, len(selected))
	for _, option := range options {
		if selected[option.Label] {
			ids = append(ids, option.ID)
		}
	}

	sort.Ints(ids)
	return ids
}

func (w *winboatApp) applySelectedMonitors(ids []int) {
	selected := make(map[int]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}

	labels := make([]string, 0, len(ids))
	for _, option := range w.monitorOptions {
		if selected[option.ID] {
			labels = append(labels, option.Label)
		}
	}

	w.monitorChecks.SetSelected(labels)
	w.refreshSelectedLabel(ids)
}

func (w *winboatApp) refreshSelectedLabel(ids []int) {
	if len(ids) == 0 {
		w.selectedLabel.SetText("None")
		return
	}

	w.selectedLabel.SetText(joinMonitorIDs(ids))
}

func (w *winboatApp) beginBusy(action string) bool {
	w.mu.Lock()
	if w.busy {
		w.mu.Unlock()
		w.appendLog("Ignored action while another action is already running.")
		return false
	}
	w.busy = true
	w.mu.Unlock()

	w.setLastAction(action + "...")
	w.setControlsBusy(true)
	w.appendLog("%s started.", action)
	return true
}

func (w *winboatApp) endBusy(message string) {
	w.mu.Lock()
	w.busy = false
	w.mu.Unlock()

	w.setControlsBusy(false)
	w.setLastAction(message)
}

func (w *winboatApp) setControlsBusy(busy bool) {
	fyne.Do(func() {
		w.mu.Lock()
		containerRunning := isRunningStatus(w.containerStatus)
		w.mu.Unlock()

		if busy {
			w.connectButton.Disable()
			w.restartConnectButton.Disable()
			w.stopButton.Disable()
			w.settingsButton.Disable()
			w.saveButton.Disable()
			w.refreshButton.Disable()
			w.refreshMonitorsButton.Disable()
			w.clearCredsButton.Disable()
		} else {
			if containerRunning {
				w.connectButton.Enable()
			} else {
				w.connectButton.Disable()
			}
			w.restartConnectButton.Enable()
			w.stopButton.Enable()
			w.settingsButton.Enable()
			w.saveButton.Enable()
			w.refreshButton.Enable()
			w.refreshMonitorsButton.Enable()
			w.clearCredsButton.Enable()
		}

		if w.traySettingsItem != nil {
			w.traySettingsItem.Disabled = busy
		}
		if w.trayConnectItem != nil {
			w.trayConnectItem.Disabled = busy || !containerRunning
		}
		if w.trayRestartItem != nil {
			w.trayRestartItem.Disabled = busy
		}
		if w.trayStopItem != nil {
			w.trayStopItem.Disabled = busy
		}
		if w.trayRefreshItem != nil {
			w.trayRefreshItem.Disabled = busy
		}
		if w.trayMenu != nil {
			w.trayMenu.Refresh()
		}
	})
}

func (w *winboatApp) setLastAction(text string) {
	fyne.Do(func() {
		w.lastActionLabel.SetText(text)
	})
}

func (w *winboatApp) appendLog(format string, args ...any) {
	line := fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	entry := logEntry{Text: line, Severity: classifyLogSeverity(line)}

	w.mu.Lock()
	w.logEntries = append(w.logEntries, entry)
	if len(w.logEntries) > maxLogLines {
		w.logEntries = append([]logEntry(nil), w.logEntries[len(w.logEntries)-maxLogLines:]...)
	}
	entries := append([]logEntry(nil), w.logEntries...)
	w.mu.Unlock()

	fyne.Do(func() {
		w.activityLog.Rows = w.buildLogRows(entries)
		w.activityLog.Refresh()
		w.activityLog.ScrollToBottom()
	})
}

func (w *winboatApp) reportError(action string, err error) {
	wrapped := fmt.Errorf("%s failed: %w", action, err)
	w.appendLog("%v", wrapped)
	fyne.Do(func() {
		dialog.ShowError(wrapped, w.window)
		w.lastActionLabel.SetText(wrapped.Error())
	})
}
