package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

const (
	hiddenLaunchArg = "--hidden"
)

type launchOptions struct {
	hidden bool
}

type autostartSettings struct {
	Enabled bool
}

func parseLaunchOptions(args []string) launchOptions {
	var options launchOptions
	for _, arg := range args {
		if arg == hiddenLaunchArg {
			options.hidden = true
		}
	}

	return options
}

func currentAutostartSettings(app fyne.App) (autostartSettings, error) {
	path, err := autostartFilePath(app)
	if err != nil {
		return autostartSettings{}, err
	}

	_, statErr := os.Stat(path)
	enabled := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return autostartSettings{}, statErr
	}

	return autostartSettings{
		Enabled: enabled,
	}, nil
}

func setAutostart(app fyne.App, settings autostartSettings) error {
	path, err := autostartFilePath(app)
	if err != nil {
		return err
	}

	if !settings.Enabled {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove autostart entry: %w", err)
		}
		return nil
	}

	exePath, err := stableExecutablePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create autostart dir: %w", err)
	}

	content := desktopEntryContents(exePath)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write autostart entry: %w", err)
	}

	return nil
}

func autostartFilePath(app fyne.App) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}

	name := app.UniqueID()
	if name == "" {
		name = "winboat-helper"
	}

	return filepath.Join(configDir, "autostart", name+".desktop"), nil
}

func stableExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}

	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}

	cleaned := filepath.Clean(exePath)
	if strings.Contains(cleaned, "go-build") {
		// return "", errors.New("autostart needs a built binary, not `go run`; build the app first and then enable Launch WinBoat when I sign in")
	}

	return cleaned, nil
}

func desktopEntryContents(executable string) string {
	execLine := strconv.Quote(executable)

	return strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Version=1.0",
		"Name=WinBoat Helper",
		"Comment=Start WinBoat Helper on login",
		"Exec=" + execLine,
		"Terminal=false",
		"StartupNotify=false",
		"X-GNOME-Autostart-enabled=true",
		"",
	}, "\n")
}
