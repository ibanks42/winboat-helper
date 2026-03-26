package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	configFolderName = "winboat-rdp"
	credentialsFile  = "credentials"
	monitorsFile     = "monitors"
)

type storedConfig struct {
	Username   string
	Password   string
	MonitorIDs []int
}

func loadStoredConfig() (storedConfig, error) {
	var cfg storedConfig

	creds, err := readShellAssignments(configFilePath(credentialsFile))
	if err != nil {
		return cfg, fmt.Errorf("load credentials: %w", err)
	}

	mons, err := readShellAssignments(configFilePath(monitorsFile))
	if err != nil {
		return cfg, fmt.Errorf("load monitors: %w", err)
	}

	cfg.Username = creds["RDP_USER"]
	cfg.Password = creds["RDP_PASS"]
	cfg.MonitorIDs, err = parseMonitorIDs(mons["RDP_MONITORS"])
	if err != nil {
		return cfg, fmt.Errorf("parse monitors: %w", err)
	}

	return cfg, nil
}

func saveStoredConfig(cfg storedConfig) error {
	if err := saveCredentials(cfg.Username, cfg.Password); err != nil {
		return err
	}

	if err := saveMonitors(cfg.MonitorIDs); err != nil {
		return err
	}

	return nil
}

func saveCredentials(username, password string) error {
	return writeShellAssignments(
		configFilePath(credentialsFile),
		[][2]string{
			{"RDP_USER", username},
			{"RDP_PASS", password},
		},
	)
}

func clearCredentials() error {
	if err := os.Remove(configFilePath(credentialsFile)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func saveMonitors(ids []int) error {
	sort.Ints(ids)

	return writeShellAssignments(
		configFilePath(monitorsFile),
		[][2]string{{"RDP_MONITORS", joinMonitorIDs(ids)}},
	)
}

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", configFolderName)
	}

	return filepath.Join(dir, configFolderName)
}

func configFilePath(name string) string {
	return filepath.Join(configDir(), name)
}

func ensureConfigDir() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	return os.Chmod(dir, 0o700)
}

func writeShellAssignments(path string, values [][2]string) error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	var lines []string
	for _, pair := range values {
		lines = append(lines, pair[0]+"="+shellQuote(pair[1]))
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return os.Chmod(path, 0o600)
}

func readShellAssignments(path string) (map[string]string, error) {
	values := map[string]string{}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return values, nil
		}

		return nil, err
	}

	for lineNo, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid assignment on line %d", lineNo+1)
		}

		key := strings.TrimSpace(parts[0])
		value, err := parseShellWord(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("parse %s on line %d: %w", key, lineNo+1, err)
		}

		values[key] = value
	}

	return values, nil
}

func parseMonitorIDs(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	ids := make([]int, 0, len(parts))
	seen := map[int]bool{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		id, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid monitor id %q", trimmed)
		}

		if seen[id] {
			continue
		}

		seen[id] = true
		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids, nil
}

func joinMonitorIDs(ids []int) string {
	if len(ids) == 0 {
		return ""
	}

	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)

	parts := make([]string, 0, len(sorted))
	for _, id := range sorted {
		parts = append(parts, strconv.Itoa(id))
	}

	return strings.Join(parts, ",")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func parseShellWord(value string) (string, error) {
	const (
		statePlain = iota
		stateSingle
		stateDouble
	)

	state := statePlain
	var out strings.Builder

	for i := 0; i < len(value); i++ {
		ch := value[i]

		switch state {
		case statePlain:
			switch ch {
			case '\\':
				i++
				if i >= len(value) {
					return "", errors.New("unfinished escape")
				}
				out.WriteByte(value[i])
			case '\'':
				state = stateSingle
			case '"':
				state = stateDouble
			default:
				out.WriteByte(ch)
			}

		case stateSingle:
			if ch == '\'' {
				state = statePlain
				continue
			}

			out.WriteByte(ch)

		case stateDouble:
			switch ch {
			case '"':
				state = statePlain
			case '\\':
				i++
				if i >= len(value) {
					return "", errors.New("unfinished escape")
				}
				out.WriteByte(value[i])
			default:
				out.WriteByte(ch)
			}
		}
	}

	if state != statePlain {
		return "", errors.New("unterminated quote")
	}

	return out.String(), nil
}
