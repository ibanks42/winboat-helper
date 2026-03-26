package main

import (
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *winboatApp) refreshStatusAsync(showErrors bool) {
	go func() {
		status, port, err := inspectWinboatState()
		if err != nil {
			w.applyStatus("Unavailable", "", time.Now())
			if showErrors {
				w.reportError("Refresh Status", err)
			}
			return
		}

		w.applyStatus(status, port, time.Now())
	}()
}

func (w *winboatApp) applyStatus(status string, port string, updatedAt time.Time) {
	w.mu.Lock()
	w.containerStatus = status
	w.mu.Unlock()

	fyne.Do(func() {
		w.containerLabel.SetText(status)
		if port == "" {
			w.portLabel.SetText("Not published")
		} else {
			w.portLabel.SetText("127.0.0.1:" + port)
		}
		w.updatedLabel.SetText(updatedAt.Format("15:04:05"))
		w.refreshSelectedLabel(w.selectedMonitorIDsFromUI())
		w.refreshPowerControl(status)
	})
}

func (w *winboatApp) refreshPowerControl(status string) {
	if isRunningStatus(status) {
		w.connectButton.Enable()
		w.stopButton.SetText("Stop VM")
		w.stopButton.SetIcon(theme.MediaStopIcon())
		w.stopButton.Importance = widget.DangerImportance
		if w.trayConnectItem != nil {
			w.trayConnectItem.Disabled = false
		}
		if w.trayStopItem != nil {
			w.trayStopItem.Label = "Stop VM"
			w.trayStopItem.Icon = theme.MediaStopIcon()
		}
	} else {
		w.connectButton.Disable()
		w.stopButton.SetText("Start VM")
		w.stopButton.SetIcon(theme.MediaPlayIcon())
		w.stopButton.Importance = widget.MediumImportance
		if w.trayConnectItem != nil {
			w.trayConnectItem.Disabled = true
		}
		if w.trayStopItem != nil {
			w.trayStopItem.Label = "Start VM"
			w.trayStopItem.Icon = theme.MediaPlayIcon()
		}
	}

	if w.trayMenu != nil {
		w.trayMenu.Refresh()
	}
}

func (w *winboatApp) waitForPort(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		port, err := getRDPPort()
		if err == nil && port != "" {
			w.appendLog("Found RDP on 127.0.0.1:%s.", port)
			return port, nil
		}

		select {
		case <-time.After(time.Second):
		case <-w.done:
			return "", errors.New("application closing")
		}
	}

	return "", fmt.Errorf("could not find RDP port for %s within %s", containerName, timeout)
}
