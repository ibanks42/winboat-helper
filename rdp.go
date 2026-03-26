package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func (w *winboatApp) launchRDP(cfg runtimeConfig, port string) error {
	backend, err := resolveRDPBackend()
	if err != nil {
		return err
	}

	args := []string{
		fmt.Sprintf("/u:%s", cfg.Username),
		"/d:.",
		"/from-stdin:force",
		fmt.Sprintf("/v:127.0.0.1:%s", port),
		"/multimon",
		fmt.Sprintf("/monitors:%s", joinMonitorIDs(cfg.MonitorIDs)),
		fmt.Sprintf("/scale:%s", defaultScale),
		"/cert:tofu",
	}

	cmd := exec.Command(backend.Command, backend.args(args...)...)
	cmd.Stdin = strings.NewReader(cfg.Password + "\n")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("capture stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("capture stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", backend.DisplayName, err)
	}

	w.appendLog("Launched %s on 127.0.0.1:%s using monitors %s.", backend.DisplayName, port, joinMonitorIDs(cfg.MonitorIDs))

	go w.streamCommandOutput(backend.DisplayName, stdout)
	go w.streamCommandOutput(backend.DisplayName, stderr)
	go func() {
		if err := cmd.Wait(); err != nil {
			w.reportError("RDP Session", fmt.Errorf("%s exited: %w", backend.DisplayName, err))
			return
		}

		w.appendLog("%s exited cleanly.", backend.DisplayName)
	}()

	return nil
}

func (w *winboatApp) streamCommandOutput(prefix string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		w.appendLog("%s: %s", prefix, line)
	}

	if err := scanner.Err(); err != nil {
		w.appendLog("%s output error: %v", prefix, err)
	}
}
