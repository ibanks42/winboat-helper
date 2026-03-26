package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func (w *winboatApp) launchRDP(cfg runtimeConfig, port string) error {
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

	cmd := exec.Command("xfreerdp", args...)
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
		return fmt.Errorf("start xfreerdp: %w", err)
	}

	w.appendLog("Launched xfreerdp on 127.0.0.1:%s using monitors %s.", port, joinMonitorIDs(cfg.MonitorIDs))

	go w.streamCommandOutput("xfreerdp", stdout)
	go w.streamCommandOutput("xfreerdp", stderr)
	go func() {
		if err := cmd.Wait(); err != nil {
			w.reportError("RDP Session", fmt.Errorf("xfreerdp exited: %w", err))
			return
		}

		w.appendLog("xfreerdp exited cleanly.")
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
