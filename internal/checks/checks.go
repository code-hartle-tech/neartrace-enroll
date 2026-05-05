package checks

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ToolStatus struct {
	Name      string
	Installed bool
	Version   string
}

type TailscaleStatus struct {
	Installed bool
	Connected bool
	IP        string
}

type Report struct {
	OS        string
	Arch      string
	Tools     []ToolStatus
	Tailscale TailscaleStatus
}

var requiredTools = []string{"tailscale", "terraform", "gh", "git"}

func Run() Report {
	r := Report{OS: runtime.GOOS, Arch: runtime.GOARCH}
	for _, name := range requiredTools {
		ts := ToolStatus{Name: name}
		if path, err := exec.LookPath(name); err == nil && path != "" {
			ts.Installed = true
			ts.Version = toolVersion(name)
		}
		r.Tools = append(r.Tools, ts)
	}
	r.Tailscale = tailscaleStatus()
	return r
}

func toolVersion(name string) string {
	args := map[string][]string{
		"tailscale": {"version"},
		"terraform": {"version", "-json"},
		"gh":        {"version"},
		"git":       {"version"},
		"brew":      {"--version"},
	}
	cmdArgs, ok := args[name]
	if !ok {
		cmdArgs = []string{"--version"}
	}
	out, err := exec.Command(name, cmdArgs...).Output()
	if err != nil {
		return "unknown"
	}
	if name == "terraform" {
		var v struct {
			TerraformVersion string `json:"terraform_version"`
		}
		if json.Unmarshal(out, &v) == nil && v.TerraformVersion != "" {
			return v.TerraformVersion
		}
	}
	return strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
}

func tailscaleStatus() TailscaleStatus {
	ts := TailscaleStatus{}
	if _, err := exec.LookPath("tailscale"); err != nil {
		return ts
	}
	ts.Installed = true
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return ts
	}
	var status struct {
		BackendState string `json:"BackendState"`
		Self         struct {
			TailscaleIPs []string `json:"TailscaleIPs"`
		} `json:"Self"`
	}
	if json.Unmarshal(out, &status) == nil {
		ts.Connected = status.BackendState == "Running"
		if len(status.Self.TailscaleIPs) > 0 {
			ts.IP = status.Self.TailscaleIPs[0]
		}
	}
	return ts
}

func (r Report) Print() {
	fmt.Printf("\nSystem: %s/%s\n\n", r.OS, r.Arch)
	for _, t := range r.Tools {
		if t.Installed {
			fmt.Printf("  ✅  %-12s %s\n", t.Name, t.Version)
		} else {
			fmt.Printf("  ❌  %-12s not installed\n", t.Name)
		}
	}
	fmt.Println()
	switch {
	case r.Tailscale.Connected:
		fmt.Printf("  ✅  tailnet       connected (%s)\n", r.Tailscale.IP)
	case r.Tailscale.Installed:
		fmt.Printf("  ⚠️   tailnet       installed but not connected\n")
	default:
		fmt.Printf("  ❌  tailnet       Tailscale not installed\n")
	}
	fmt.Println()
}

func (r Report) AllGreen() bool {
	for _, t := range r.Tools {
		if !t.Installed {
			return false
		}
	}
	return r.Tailscale.Connected
}
