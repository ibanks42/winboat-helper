package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func detectMonitors() ([]monitorOption, rdpBackend, error) {
	backend, err := resolveRDPBackend()
	if err != nil {
		return nil, rdpBackend{}, err
	}

	output, err := runCommand(context.Background(), 15*time.Second, backend.Command, backend.args("/list:monitor")...)
	if err != nil {
		return nil, rdpBackend{}, fmt.Errorf("list monitors with %s: %w", backend.DisplayName, err)
	}

	var options []monitorOption
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		match := monitorLinePattern.FindStringSubmatch(line)
		if len(match) != 5 {
			continue
		}

		id := mustAtoi(match[1])
		label := fmt.Sprintf("%d - %s (%s,%s)", id, match[2], signedCoordinate(match[3]), signedCoordinate(match[4]))
		options = append(options, monitorOption{ID: id, Label: label})
	}

	if len(options) == 0 {
		return nil, rdpBackend{}, fmt.Errorf("%s did not report any monitors", backend.DisplayName)
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].ID < options[j].ID
	})

	return options, backend, nil
}

func inspectWinboatState() (status string, port string, err error) {
	status, err = runCommand(context.Background(), 15*time.Second, "docker", "inspect", "-f", "{{.State.Status}}", containerName)
	if err != nil {
		return "", "", fmt.Errorf("inspect container: %w", err)
	}

	status = strings.TrimSpace(status)
	port, err = getRDPPort()
	if err != nil {
		return status, "", nil
	}

	return status, port, nil
}

func getRDPPort() (string, error) {
	output, err := runCommand(context.Background(), 10*time.Second, "docker", "port", containerName, "3389/tcp")
	if err != nil {
		return "", err
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "", errors.New("no published RDP port")
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected docker port output: %q", trimmed)
	}

	return strings.TrimSpace(parts[len(parts)-1]), nil
}

func runCommand(parent context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		return trimmed, fmt.Errorf("%s timed out", name)
	}

	if err != nil {
		if trimmed == "" {
			return "", err
		}

		return trimmed, fmt.Errorf("%w: %s", err, trimmed)
	}

	return trimmed, nil
}

func allMonitorIDs(options []monitorOption) []int {
	ids := make([]int, 0, len(options))
	for _, option := range options {
		ids = append(ids, option.ID)
	}

	return ids
}

func signedCoordinate(value string) string {
	if strings.HasPrefix(value, "-") || strings.HasPrefix(value, "+") {
		return value
	}

	return "+" + value
}

func mustAtoi(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}

	return parsed
}

func isRunningStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "running")
}
