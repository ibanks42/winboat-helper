package main

import (
	"errors"
	"fmt"
	"os"
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
		Scale:      normalizeScale(w.scaleSelect.Selected),
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

func (w *winboatApp) backendMonitorIDsForSelected(ids []int) []int {
	selected := make(map[int]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}

	w.mu.Lock()
	options := append([]monitorOption(nil), w.monitorOptions...)
	w.mu.Unlock()

	backendIDs := make([]int, 0, len(ids))
	for _, option := range options {
		if selected[option.ID] {
			backendIDs = append(backendIDs, option.BackendID)
		}
	}

	sort.Ints(backendIDs)
	return backendIDs
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

		trayChanged := false

		if w.connectButton != nil {
			if busy || !containerRunning {
				w.connectButton.Disable()
			} else {
				w.connectButton.Enable()
			}
		}
		if w.restartConnectButton != nil {
			if busy {
				w.restartConnectButton.Disable()
			} else {
				w.restartConnectButton.Enable()
			}
		}
		if w.stopButton != nil {
			if busy {
				w.stopButton.Disable()
			} else {
				w.stopButton.Enable()
			}
		}
		if w.settingsButton != nil {
			if busy {
				w.settingsButton.Disable()
			} else {
				w.settingsButton.Enable()
			}
		}
		if w.saveButton != nil {
			if busy {
				w.saveButton.Disable()
			} else {
				w.saveButton.Enable()
			}
		}
		if w.refreshButton != nil {
			if busy {
				w.refreshButton.Disable()
			} else {
				w.refreshButton.Enable()
			}
		}
		if w.refreshMonitorsButton != nil {
			if busy {
				w.refreshMonitorsButton.Disable()
			} else {
				w.refreshMonitorsButton.Enable()
			}
		}
		if w.clearCredsButton != nil {
			if busy {
				w.clearCredsButton.Disable()
			} else {
				w.clearCredsButton.Enable()
			}
		}

		if w.traySettingsItem != nil {
			trayChanged = updateTrayItemDisabled(w.traySettingsItem, busy) || trayChanged
		}
		if w.trayConnectItem != nil {
			trayChanged = updateTrayItemDisabled(w.trayConnectItem, busy || !containerRunning) || trayChanged
		}
		if w.trayRestartItem != nil {
			trayChanged = updateTrayItemDisabled(w.trayRestartItem, busy) || trayChanged
		}
		if w.trayStopItem != nil {
			trayChanged = updateTrayItemDisabled(w.trayStopItem, busy) || trayChanged
		}
		if trayChanged && w.trayMenu != nil {
			w.trayMenu.Refresh()
		}
	})
}

func updateTrayItemDisabled(item *fyne.MenuItem, disabled bool) bool {
	if item == nil || item.Disabled == disabled {
		return false
	}

	item.Disabled = disabled
	return true
}

func (w *winboatApp) setLastAction(text string) {
	fyne.Do(func() {
		w.lastActionLabel.SetText(text)
	})
}

func (w *winboatApp) appendLog(format string, args ...any) {
	line := fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	entry := logEntry{Text: line, Severity: classifyLogSeverity(line)}
	w.writeLogFileLine(line)

	w.mu.Lock()
	w.logEntries = append(w.logEntries, entry)
	if len(w.logEntries) > maxLogLines {
		w.logEntries = append([]logEntry(nil), w.logEntries[len(w.logEntries)-maxLogLines:]...)
	}
	text := w.joinLogEntriesLocked()
	w.mu.Unlock()

	fyne.Do(func() {
		w.setActivityLogText(text)
	})
}

func (w *winboatApp) logText() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.joinLogEntriesLocked()
}

func (w *winboatApp) joinLogEntriesLocked() string {
	lines := make([]string, 0, len(w.logEntries))
	for _, entry := range w.logEntries {
		lines = append(lines, entry.Text)
	}

	return strings.Join(lines, "\n")
}

func (w *winboatApp) setActivityLogText(text string) {
	w.mu.Lock()
	w.applyingLogText = true
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.applyingLogText = false
		w.mu.Unlock()
	}()

	w.activityLog.SetText(text)
	lineCount := 0
	lastLineLen := 0
	if text != "" {
		lines := strings.Split(text, "\n")
		lineCount = len(lines) - 1
		lastLineLen = len([]rune(lines[len(lines)-1]))
	}
	w.activityLog.CursorRow = lineCount
	w.activityLog.CursorColumn = lastLineLen
	w.activityLog.Refresh()
}

func (w *winboatApp) writeLogFileLine(line string) {
	if err := ensureConfigDir(); err != nil {
		return
	}

	file, err := os.OpenFile(configFilePath(logsFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer file.Close()

	_, _ = file.WriteString(line + "\n")
}

func (w *winboatApp) reportError(action string, err error) {
	wrapped := fmt.Errorf("%s failed: %w", action, err)
	w.appendLog("%v", wrapped)
	fyne.Do(func() {
		w.mu.Lock()
		settingsShown := w.settingsShown
		logShown := w.logShown
		w.mu.Unlock()

		switch {
		case logShown && w.logWindow != nil:
			dialog.ShowError(wrapped, w.logWindow)
		case settingsShown && w.settingsWindow != nil:
			dialog.ShowError(wrapped, w.settingsWindow)
		}
		w.lastActionLabel.SetText(wrapped.Error())
	})
}
