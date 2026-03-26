package main

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

const flatpakFreeRDPAppID = "com.freerdp.FreeRDP"

type rdpBackend struct {
	DisplayName string
	Command     string
	PrefixArgs  []string
}

func (r rdpBackend) args(extra ...string) []string {
	args := append([]string(nil), r.PrefixArgs...)
	return append(args, extra...)
}

func resolveRDPBackend() (rdpBackend, error) {
	if _, err := exec.LookPath("xfreerdp"); err == nil {
		return rdpBackend{DisplayName: "xfreerdp", Command: "xfreerdp"}, nil
	}

	if _, err := exec.LookPath("xfreerdp3"); err == nil {
		return rdpBackend{DisplayName: "xfreerdp3", Command: "xfreerdp3"}, nil
	}

	if _, err := exec.LookPath("flatpak"); err == nil {
		flatpakBackend, err := resolveFlatpakRDPBackend()
		if err == nil {
			return flatpakBackend, nil
		}
	}

	return rdpBackend{}, errors.New("could not find FreeRDP; install `xfreerdp`, `xfreerdp3`, or Flatpak `com.freerdp.FreeRDP`")
}

func resolveFlatpakRDPBackend() (rdpBackend, error) {
	if _, err := runCommand(context.Background(), 10*time.Second, "flatpak", "info", flatpakFreeRDPAppID); err != nil {
		return rdpBackend{}, err
	}

	candidates := []rdpBackend{
		{
			DisplayName: "Flatpak FreeRDP (xfreerdp)",
			Command:     "flatpak",
			PrefixArgs:  []string{"run", "--command=xfreerdp", flatpakFreeRDPAppID},
		},
		{
			DisplayName: "Flatpak FreeRDP (xfreerdp3)",
			Command:     "flatpak",
			PrefixArgs:  []string{"run", "--command=xfreerdp3", flatpakFreeRDPAppID},
		},
	}

	for _, candidate := range candidates {
		if _, err := runCommand(context.Background(), 15*time.Second, candidate.Command, candidate.args("/version")...); err == nil {
			return candidate, nil
		}
	}

	return rdpBackend{}, errors.New("found Flatpak FreeRDP, but could not locate `xfreerdp` or `xfreerdp3` inside the Flatpak runtime")
}
