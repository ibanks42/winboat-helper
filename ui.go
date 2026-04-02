package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *winboatApp) buildUI() {
	w.usernameEntry = widget.NewEntry()
	w.usernameEntry.SetPlaceHolder("Windows username")
	w.usernameEntry.SetIcon(theme.AccountIcon())

	w.passwordEntry = widget.NewPasswordEntry()
	w.passwordEntry.SetPlaceHolder("Windows password")
	w.passwordEntry.SetIcon(theme.LoginIcon())

	w.scaleSelect = widget.NewSelect(supportedScales, nil)
	w.scaleSelect.SetSelected(defaultScale)

	w.monitorChecks = widget.NewCheckGroup(nil, nil)
	w.monitorHint = widget.NewLabel("Pick the monitors to use for the RDP session.")
	w.monitorHint.Wrapping = fyne.TextWrapWord

	w.containerLabel = widget.NewLabel("Checking...")
	w.portLabel = widget.NewLabel("Checking...")
	w.updatedLabel = widget.NewLabel("Waiting for first refresh...")
	w.selectedLabel = widget.NewLabel("None")
	w.selectedLabel.Wrapping = fyne.TextWrapWord
	w.lastActionLabel = widget.NewLabel("Ready")
	w.lastActionLabel.Wrapping = fyne.TextWrapWord

	w.activityLog = widget.NewMultiLineEntry()
	w.activityLog.Wrapping = fyne.TextWrapOff
	w.activityLog.Scroll = fyne.ScrollVerticalOnly
	w.activityLog.TextStyle = fyne.TextStyle{Monospace: true}
	w.activityLog.SetMinRowsVisible(14)
	w.activityLog.OnChanged = func(text string) {
		w.mu.Lock()
		applying := w.applyingLogText
		w.mu.Unlock()
		if applying {
			return
		}

		logs := w.logText()
		if text == logs {
			return
		}

		fyne.Do(func() {
			w.setActivityLogText(logs)
		})
	}

	w.launchAtLoginCheck = widget.NewCheck("Launch WinBoat when I sign in", nil)

	w.saveButton = widget.NewButtonWithIcon("Save Settings", theme.DocumentSaveIcon(), func() {
		w.saveSettings()
	})

	w.refreshMonitorsButton = widget.NewButtonWithIcon("Refresh Monitors", theme.ViewRefreshIcon(), func() {
		w.reloadMonitors(true)
	})

	w.clearCredsButton = widget.NewButtonWithIcon("Clear Stored Credentials", theme.DeleteIcon(), func() {
		w.clearStoredCredentials()
	})

	w.copyLogsButton = widget.NewButtonWithIcon("Copy Logs", theme.ContentCopyIcon(), func() {
		logs := w.logText()
		if strings.TrimSpace(logs) == "" {
			w.appendLog("No logs to copy yet.")
			return
		}

		w.app.Clipboard().SetContent(logs)
		w.appendLog("Copied %d log line(s) to the clipboard. Log file: %s.", strings.Count(logs, "\n")+1, configFilePath(logsFile))
	})

	w.buildLogWindow()
	w.buildSettingsWindow()
}

func (w *winboatApp) buildLogWindow() {
	statusForm := widget.NewForm(
		widget.NewFormItem("Container", w.containerLabel),
		widget.NewFormItem("RDP Port", w.portLabel),
		widget.NewFormItem("Updated", w.updatedLabel),
		widget.NewFormItem("Selected Monitors", w.selectedLabel),
		widget.NewFormItem("Last Action", w.lastActionLabel),
	)

	content := container.NewBorder(
		widget.NewCard("Status", "Live container state, port mapping, and current monitor selection.", statusForm),
		nil,
		nil,
		nil,
		widget.NewCard("Activity", "Recent events from the app and the RDP process.", container.NewBorder(
			container.NewHBox(layout.NewSpacer(), w.copyLogsButton),
			nil,
			nil,
			nil,
			container.NewStack(w.activityLog),
		)),
	)

	w.logWindow = w.app.NewWindow("WinBoat Log")
	w.logWindow.SetIcon(appIcon)
	w.logWindow.Resize(fyne.NewSize(760, 720))
	w.logWindow.SetCloseIntercept(func() {
		w.mu.Lock()
		w.logShown = false
		w.mu.Unlock()
		w.logWindow.Hide()
		w.notifyHideToTray()
	})
	w.logWindow.SetContent(container.NewPadded(content))
}

func (w *winboatApp) buildSettingsWindow() {
	settingsForm := widget.NewForm(
		widget.NewFormItem("Username", w.usernameEntry),
		widget.NewFormItem("Password", w.passwordEntry),
		widget.NewFormItem("Scale", w.scaleSelect),
	)

	startupSection := widget.NewCard("Startup", "Create or remove a login autostart entry for this user.", container.NewVBox(
		w.launchAtLoginCheck,
	))

	content := container.NewVBox(
		widget.NewLabel("Stored locally in ~/.config/winboat-rdp with 0600 file permissions."),
		settingsForm,
		widget.NewSeparator(),
		startupSection,
		widget.NewSeparator(),
		w.monitorHint,
		w.monitorChecks,
		container.NewHBox(w.refreshMonitorsButton, layout.NewSpacer(), w.clearCredsButton, w.saveButton),
	)

	w.settingsWindow = w.app.NewWindow("WinBoat Settings")
	w.settingsWindow.SetIcon(appIcon)
	w.settingsWindow.Resize(fyne.NewSize(760, 620))
	w.settingsWindow.SetCloseIntercept(func() {
		w.mu.Lock()
		w.settingsShown = false
		w.mu.Unlock()
		w.settingsWindow.Hide()
		w.notifyHideToTray()
	})
	w.settingsWindow.SetContent(container.NewPadded(content))
}

func (w *winboatApp) installTray() {
	desk, ok := w.app.(desktop.App)
	if !ok {
		w.appendLog("System tray is not available in this Fyne driver.")
		return
	}

	w.traySettingsItem = fyne.NewMenuItemWithIcon("Settings", theme.SettingsIcon(), func() {
		w.showSettingsWindow()
	})
	w.trayLogItem = fyne.NewMenuItemWithIcon("Logs", theme.FileTextIcon(), func() {
		w.showLogWindow()
	})
	w.trayConnectItem = fyne.NewMenuItemWithIcon("Connect", theme.MediaPlayIcon(), func() {
		w.connect()
	})
	w.trayRestartItem = fyne.NewMenuItemWithIcon("Restart and Connect", theme.MediaReplayIcon(), func() {
		w.restartAndConnect()
	})
	w.trayStopItem = fyne.NewMenuItemWithIcon("Stop VM", theme.MediaStopIcon(), func() {
		w.toggleVM()
	})
	w.trayQuitItem = fyne.NewMenuItemWithIcon("Quit", theme.LogoutIcon(), func() {
		w.signalDone()
		w.app.Quit()
	})
	w.trayQuitItem.IsQuit = true

	w.trayMenu = fyne.NewMenu("WinBoat Helper",
		w.traySettingsItem,
		w.trayLogItem,
		fyne.NewMenuItemSeparator(),
		w.trayConnectItem,
		w.trayRestartItem,
		w.trayStopItem,
		fyne.NewMenuItemSeparator(),
		w.trayQuitItem,
	)

	desk.SetSystemTrayMenu(w.trayMenu)
	desk.SetSystemTrayIcon(appIcon)
}

func (w *winboatApp) showSettingsWindow() {
	if w.settingsWindow == nil {
		return
	}

	fyne.Do(func() {
		w.mu.Lock()
		w.settingsShown = true
		w.mu.Unlock()
		w.settingsWindow.Show()
		w.settingsWindow.RequestFocus()
	})
}

func (w *winboatApp) showLogWindow() {
	if w.logWindow == nil {
		return
	}

	fyne.Do(func() {
		w.mu.Lock()
		w.logShown = true
		w.mu.Unlock()
		w.logWindow.Show()
		w.logWindow.RequestFocus()
	})
}
