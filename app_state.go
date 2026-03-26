package main

import (
	"regexp"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	containerName   = "WinBoat"
	defaultScale    = "180"
	portWaitTimeout = 60 * time.Second
	statusPollEvery = 3 * time.Second
	maxLogLines     = 200
)

var supportedScales = []string{"100", "140", "180"}

var monitorLinePattern = regexp.MustCompile(`\[(\d+)\]\s+(\d+x\d+)\s+\+(-?\d+)\+(-?\d+)`)

type monitorOption struct {
	ID    int
	Label string
}

type runtimeConfig struct {
	Username   string
	Password   string
	MonitorIDs []int
	Scale      string
}

type winboatApp struct {
	app    fyne.App
	window fyne.Window

	usernameEntry   *widget.Entry
	passwordEntry   *widget.Entry
	scaleSelect     *widget.Select
	monitorChecks   *widget.CheckGroup
	monitorHint     *widget.Label
	containerLabel  *widget.Label
	portLabel       *widget.Label
	updatedLabel    *widget.Label
	selectedLabel   *widget.Label
	lastActionLabel *widget.Label
	activityLog     *widget.TextGrid

	connectButton         *widget.Button
	restartConnectButton  *widget.Button
	stopButton            *widget.Button
	settingsButton        *widget.Button
	launchAtLoginCheck    *widget.Check
	startHiddenCheck      *widget.Check
	saveButton            *widget.Button
	refreshButton         *widget.Button
	refreshMonitorsButton *widget.Button
	clearCredsButton      *widget.Button

	trayMenu         *fyne.Menu
	trayShowItem     *fyne.MenuItem
	traySettingsItem *fyne.MenuItem
	trayConnectItem  *fyne.MenuItem
	trayRestartItem  *fyne.MenuItem
	trayStopItem     *fyne.MenuItem
	trayRefreshItem  *fyne.MenuItem
	trayQuitItem     *fyne.MenuItem
	settingsDialog   *dialog.CustomDialog

	mu                sync.Mutex
	quitOnce          sync.Once
	busy              bool
	containerStatus   string
	monitorOptions    []monitorOption
	preferredMonitors []int
	labelToMonitor    map[string]int
	logEntries        []logEntry
	done              chan struct{}
}

func newWinboatApp(a fyne.App) *winboatApp {
	state := &winboatApp{
		app:            a,
		labelToMonitor: map[string]int{},
		done:           make(chan struct{}),
	}

	state.buildUI()
	state.installTray()
	state.loadInitialState()
	state.startStatusPolling()

	return state
}
