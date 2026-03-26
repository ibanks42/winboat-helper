package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Icon.png
var appIconContent []byte

var appIcon = &fyne.StaticResource{
	StaticName:    "Icon.png",
	StaticContent: appIconContent,
}
