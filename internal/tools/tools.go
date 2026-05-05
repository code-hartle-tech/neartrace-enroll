package tools

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Tool struct {
	Name        string
	BrewPackage string
	WingetID    string
	AptPackage  string
}

var Required = []Tool{
	{Name: "tailscale", BrewPackage: "tailscale", WingetID: "tailscale.tailscale", AptPackage: "tailscale"},
	{Name: "terraform", BrewPackage: "terraform", WingetID: "Hashicorp.Terraform", AptPackage: "terraform"},
	{Name: "gh", BrewPackage: "gh", WingetID: "GitHub.cli", AptPackage: "gh"},
	{Name: "git", BrewPackage: "git", WingetID: "Git.Git", AptPackage: "git"},
}

// InstallMissing installs any tool not found on PATH.
func InstallMissing() error {
	for _, t := range Required {
		if _, err := exec.LookPath(t.Name); err == nil {
			fmt.Printf("  ✅  %-12s already installed\n", t.Name)
			continue
		}
		fmt.Printf("  →   %-12s installing...\n", t.Name)
		if err := install(t); err != nil {
			return fmt.Errorf("install %s: %w", t.Name, err)
		}
		fmt.Printf("  ✅  %-12s installed\n", t.Name)
	}
	return nil
}

func install(t Tool) error {
	switch runtime.GOOS {
	case "darwin":
		return brewInstall(t.BrewPackage)
	case "windows":
		return wingetInstall(t.WingetID)
	case "linux":
		if pm := detectLinuxPM(); pm == "apt" {
			return aptInstall(t.AptPackage)
		}
		return brewInstall(t.BrewPackage) // Homebrew on Linux fallback
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func brewInstall(pkg string) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found — install from https://brew.sh first")
	}
	cmd := exec.Command("brew", "install", pkg)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func wingetInstall(id string) error {
	cmd := exec.Command("winget", "install", "--id", id, "--silent", "--accept-package-agreements", "--accept-source-agreements")
	return cmd.Run()
}

func aptInstall(pkg string) error {
	if err := exec.Command("sudo", "apt-get", "update", "-qq").Run(); err != nil {
		return err
	}
	return exec.Command("sudo", "apt-get", "install", "-y", pkg).Run()
}

func detectLinuxPM() string {
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "apt"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf"
	}
	return "brew"
}
