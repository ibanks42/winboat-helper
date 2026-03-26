package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *winboatApp) buildUI() {
	w.window = w.app.NewWindow("WinBoat Helper")
	w.window.SetIcon(appIcon)
	w.window.Resize(fyne.NewSize(760, 720))
	w.window.SetCloseIntercept(func() {
		w.window.Hide()
		w.notifyHideToTray()
	})

	w.usernameEntry = widget.NewEntry()
	w.usernameEntry.SetPlaceHolder("Windows username")
	w.usernameEntry.SetIcon(theme.AccountIcon())

	w.passwordEntry = widget.NewPasswordEntry()
	w.passwordEntry.SetPlaceHolder("Windows password")
	w.passwordEntry.SetIcon(theme.LoginIcon())

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

	w.activityLog = widget.NewTextGrid()
	w.activityLog.Scroll = fyne.ScrollVerticalOnly

	w.connectButton = widget.NewButtonWithIcon("Connect", theme.MediaPlayIcon(), func() {
		w.connect(false)
	})
	w.connectButton.Importance = widget.HighImportance

	w.restartConnectButton = widget.NewButtonWithIcon("Restart + Connect", theme.MediaReplayIcon(), func() {
		w.connect(true)
	})

	w.stopButton = widget.NewButtonWithIcon("Stop VM", theme.MediaStopIcon(), func() {
		w.toggleVM()
	})
	w.stopButton.Importance = widget.DangerImportance

	w.settingsButton = widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		w.showSettingsWindow()
	})

	w.launchAtLoginCheck = widget.NewCheck("Launch WinBoat when I sign in", nil)
	w.startHiddenCheck = widget.NewCheck("Start hidden in the tray on auto-launch", nil)
	w.launchAtLoginCheck.OnChanged = func(bool) {
		w.syncStartupControls()
	}

	w.saveButton = widget.NewButtonWithIcon("Save Settings", theme.DocumentSaveIcon(), func() {
		w.saveSettings()
	})

	w.refreshButton = widget.NewButtonWithIcon("Refresh Status", theme.ViewRefreshIcon(), func() {
		w.refreshStatusAsync(true)
	})

	w.refreshMonitorsButton = widget.NewButtonWithIcon("Refresh Monitors", theme.ViewRefreshIcon(), func() {
		w.reloadMonitors(true)
	})

	w.clearCredsButton = widget.NewButtonWithIcon("Clear Stored Credentials", theme.DeleteIcon(), func() {
		w.clearStoredCredentials()
	})

	statusForm := widget.NewForm(
		widget.NewFormItem("Container", w.containerLabel),
		widget.NewFormItem("RDP Port", w.portLabel),
		widget.NewFormItem("Updated", w.updatedLabel),
		widget.NewFormItem("Selected Monitors", w.selectedLabel),
		widget.NewFormItem("Last Action", w.lastActionLabel),
	)

	contentTop := container.NewVBox(
		widget.NewCard("Status", "Live container state, port mapping, and current monitor selection.", container.NewVBox(
			statusForm,
			container.NewHBox(w.refreshButton, layout.NewSpacer(), w.settingsButton),
		)),
		widget.NewCard("Actions", "Launch RDP directly or manage the WinBoat container.", container.NewHBox(
			w.connectButton,
			w.restartConnectButton,
			w.stopButton,
		)),
	)

	content := container.NewBorder(
		contentTop,
		nil,
		nil,
		nil,
		widget.NewCard("Activity", "Recent events from the app and the RDP process.", container.NewStack(w.activityLog)),
	)

	w.window.SetContent(container.NewPadded(content))
	if desk, ok := w.app.(desktop.App); ok {
		desk.SetSystemTrayWindow(w.window)
	}
	w.window.SetMaster()
}

func (w *winboatApp) installTray() {
	desk, ok := w.app.(desktop.App)
	if !ok {
		w.appendLog("System tray is not available in this Fyne driver.")
		return
	}

	w.trayShowItem = fyne.NewMenuItemWithIcon("Show Window", theme.VisibilityIcon(), func() {
		w.showWindow()
	})
	w.traySettingsItem = fyne.NewMenuItemWithIcon("Settings", theme.SettingsIcon(), func() {
		w.showSettingsWindow()
	})
	w.trayConnectItem = fyne.NewMenuItemWithIcon("Connect", theme.MediaPlayIcon(), func() {
		w.connect(false)
	})
	w.trayRestartItem = fyne.NewMenuItemWithIcon("Restart + Connect", theme.MediaReplayIcon(), func() {
		w.connect(true)
	})
	w.trayStopItem = fyne.NewMenuItemWithIcon("Stop VM", theme.MediaStopIcon(), func() {
		w.toggleVM()
	})
	w.trayRefreshItem = fyne.NewMenuItemWithIcon("Refresh Status", theme.ViewRefreshIcon(), func() {
		w.refreshStatusAsync(true)
	})
	w.trayQuitItem = fyne.NewMenuItemWithIcon("Quit", theme.LogoutIcon(), func() {
		w.signalDone()
		w.app.Quit()
	})
	w.trayQuitItem.IsQuit = true

	w.trayMenu = fyne.NewMenu("WinBoat Helper",
		w.trayShowItem,
		w.traySettingsItem,
		fyne.NewMenuItemSeparator(),
		w.trayConnectItem,
		w.trayRestartItem,
		w.trayStopItem,
		w.trayRefreshItem,
		fyne.NewMenuItemSeparator(),
		w.trayQuitItem,
	)

	desk.SetSystemTrayMenu(w.trayMenu)
	desk.SetSystemTrayIcon(appIcon)
	desk.SetSystemTrayWindow(w.window)
}

func (w *winboatApp) showWindow() {
	if w.window == nil {
		return
	}

	fyne.Do(func() {
		w.window.Show()
		w.window.RequestFocus()
	})
}

func (w *winboatApp) buildSettingsWindow() {
	settingsForm := widget.NewForm(
		widget.NewFormItem("Username", w.usernameEntry),
		widget.NewFormItem("Password", w.passwordEntry),
	)

	startupSection := widget.NewCard("Startup", "Create or remove a login autostart entry for this user.", container.NewVBox(
		w.launchAtLoginCheck,
		w.startHiddenCheck,
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

	w.settingsDialog = dialog.NewCustom("WinBoat Helper Settings", "Close", container.NewPadded(content), w.window)
	w.settingsDialog.Resize(fyne.NewSize(760, 620))
	w.settingsDialog.SetIcon(appIcon)

	w.syncStartupControls()
}

func (w *winboatApp) showSettingsWindow() {
	if w.settingsDialog == nil {
		w.buildSettingsWindow()
	}

	w.settingsDialog.Show()
}
