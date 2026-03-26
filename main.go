package main

import (
	"os"

	"fyne.io/fyne/v2/app"
)

func main() {
	options := parseLaunchOptions(os.Args[1:])

	a := app.NewWithID("dev.ibanks.winboat-helper")
	a.SetIcon(appIcon)

	ui := newWinboatApp(a)
	if options.hidden && ui.trayMenu != nil {
		ui.appendLog("Started hidden.")
	} else {
		ui.window.Show()
	}

	a.Run()
}
