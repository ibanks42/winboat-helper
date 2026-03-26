package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type logSeverity int

const (
	logSeverityInfo logSeverity = iota
	logSeverityWarn
	logSeverityError
)

type logEntry struct {
	Text     string
	Severity logSeverity
}

func classifyLogSeverity(text string) logSeverity {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "failed"), strings.Contains(lower, "error"), strings.Contains(lower, "panic"):
		return logSeverityError
	case strings.Contains(lower, "warn"), strings.Contains(lower, "timeout"), strings.Contains(lower, "unavailable"):
		return logSeverityWarn
	default:
		return logSeverityInfo
	}
}

func (w *winboatApp) buildLogRows(entries []logEntry) []widget.TextGridRow {
	rows := make([]widget.TextGridRow, 0, len(entries))

	for _, entry := range entries {
		style := w.logStyleForSeverity(entry.Severity)
		cells := make([]widget.TextGridCell, 0, len(entry.Text))
		for _, r := range entry.Text {
			cells = append(cells, widget.TextGridCell{Rune: r, Style: style})
		}
		rows = append(rows, widget.TextGridRow{Cells: cells, Style: style})
	}

	if len(rows) == 0 {
		rows = append(rows, widget.TextGridRow{})
	}

	return rows
}

func (w *winboatApp) logStyleForSeverity(severity logSeverity) widget.TextGridStyle {
	colorName := theme.ColorNameForeground
	switch severity {
	case logSeverityError:
		colorName = theme.ColorNameError
	case logSeverityWarn:
		colorName = theme.ColorNameWarning
	}

	return &widget.CustomTextGridStyle{
		TextStyle: fyne.TextStyle{Monospace: true},
		FGColor:   theme.Color(colorName),
		BGColor:   color.Transparent,
	}
}
