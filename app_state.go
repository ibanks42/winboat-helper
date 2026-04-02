package main

import (
	"regexp"
	"sync"
	"time"

	"fyne.io/fyne/v2"
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

var monitorLinePattern = regexp.MustCompile(`\[(\d+)\].*?(\d+x\d+)\s+\+(-?\d+)\+(-?\d+)`)

type monitorOption struct {
	ID        int
	BackendID int
	Label     string
	Width     int
	Height    int
	X         int
	Y         int
}

type runtimeConfig struct {
	Username   string
	Password   string
	MonitorIDs []int
	Scale      string
}

type winboatApp struct {
	app fyne.App

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
	activityLog     *widget.Entry

	connectButton         *widget.Button
	restartConnectButton  *widget.Button
	stopButton            *widget.Button
	settingsButton        *widget.Button
	launchAtLoginCheck    *widget.Check
	saveButton            *widget.Button
	refreshButton         *widget.Button
	refreshMonitorsButton *widget.Button
	clearCredsButton      *widget.Button
	copyLogsButton        *widget.Button

	trayMenu         *fyne.Menu
	trayLogItem      *fyne.MenuItem
	traySettingsItem *fyne.MenuItem
	trayConnectItem  *fyne.MenuItem
	trayRestartItem  *fyne.MenuItem
	trayStopItem     *fyne.MenuItem
	trayQuitItem     *fyne.MenuItem
	settingsWindow   fyne.Window
	logWindow        fyne.Window

	mu                sync.Mutex
	quitOnce          sync.Once
	busy              bool
	containerStatus   string
	monitorOptions    []monitorOption
	preferredMonitors []int
	labelToMonitor    map[string]int
	logEntries        []logEntry
	applyingLogText   bool
	settingsShown     bool
	logShown          bool
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
