package main

import (
	"os"

	"fyne.io/fyne/v2/app"
)

func main() {
	_ = parseLaunchOptions(os.Args[1:])

	a := app.NewWithID("dev.ibanks.winboat-helper")
	a.SetIcon(appIcon)

	newWinboatApp(a)

	a.Run()
}
