package tailscale

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Status struct {
	Connected bool
	IP        string
	Hostname  string
}

// Enroll runs `tailscale up`, opening the browser for Google Workspace auth.
// Blocks until the device appears on the tailnet (up to 3 minutes).
func Enroll() error {
	if s, _ := GetStatus(); s.Connected {
		fmt.Printf("  ✅  Already enrolled (IP: %s)\n", s.IP)
		return nil
	}

	fmt.Println("  →   Opening browser for @hartle.tech Google sign-in...")
	cmd := exec.Command("tailscale", "up", "--accept-routes")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("tailscale up: %w", err)
	}

	fmt.Print("  →   Waiting for enrollment")
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		fmt.Print(".")
		if s, err := GetStatus(); err == nil && s.Connected {
			fmt.Printf("\n  ✅  Enrolled — IP: %s\n", s.IP)
			return nil
		}
	}
	fmt.Println()
	return fmt.Errorf("timed out waiting for Tailscale enrollment — check the browser window")
}

func GetStatus() (Status, error) {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return Status{}, err
	}
	var raw struct {
		BackendState string `json:"BackendState"`
		Self         struct {
			DNSName      string   `json:"DNSName"`
			TailscaleIPs []string `json:"TailscaleIPs"`
		} `json:"Self"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return Status{}, err
	}
	s := Status{
		Connected: raw.BackendState == "Running",
		Hostname:  raw.Self.DNSName,
	}
	if len(raw.Self.TailscaleIPs) > 0 {
		s.IP = raw.Self.TailscaleIPs[0]
	}
	return s, nil
}
