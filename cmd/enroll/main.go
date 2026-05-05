package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tech.hartle/neartrace-enroll/internal/checks"
	"tech.hartle/neartrace-enroll/internal/repos"
	"tech.hartle/neartrace-enroll/internal/tailscale"
	"tech.hartle/neartrace-enroll/internal/tokens"
	"tech.hartle/neartrace-enroll/internal/tools"
)

var rootCmd = &cobra.Command{
	Use:   "enroll",
	Short: "HARTLE.TECH machine enrollment kit",
	Long: `Enroll this machine onto the HARTLE.TECH tailnet and install all dev tools.
Sign in with your @hartle.tech Google account when prompted.`,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Full onboarding — install tools, enroll Tailscale, seed tokens, clone repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		githubToken, _ := cmd.Flags().GetString("github-token")
		baseDir, _ := cmd.Flags().GetString("projects-dir")

		banner("HARTLE.TECH Enrollment")

		step("Installing dev tools")
		if err := tools.InstallMissing(); err != nil {
			return err
		}

		step("Enrolling on tailnet")
		if err := tailscale.Enroll(); err != nil {
			return err
		}

		step("Seeding credentials file")
		if err := tokens.Seed(); err != nil {
			return err
		}

		if githubToken != "" {
			step("Cloning repositories")
			if err := repos.CloneAll(baseDir, githubToken); err != nil {
				return err
			}
		} else {
			fmt.Println("\n  ℹ️  Pass --github-token to also clone all repositories.")
		}

		step("Final health check")
		r := checks.Run()
		r.Print()

		if r.AllGreen() {
			banner("All done ✅")
			fmt.Printf("  Fill in empty values in %s\n", tokens.File())
			fmt.Println("  Then: source ~/.config/hartle.tech/tokens && terraform apply")
		} else {
			fmt.Println("  ⚠️  Some checks failed — review above and re-run.")
		}
		return nil
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Print system health — installed tools, Tailscale status",
	Run: func(cmd *cobra.Command, args []string) {
		r := checks.Run()
		r.Print()
		if !r.AllGreen() {
			os.Exit(1)
		}
	},
}

var tailscaleCmd = &cobra.Command{
	Use:   "tailscale",
	Short: "Enroll this machine on the tailnet only",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tailscale.Enroll()
	},
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Install missing dev tools only",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tools.InstallMissing()
	},
}

func init() {
	runCmd.Flags().String("github-token", "", "GitHub PAT for cloning private repos")
	runCmd.Flags().String("projects-dir", os.ExpandEnv("$HOME/Projects"), "Directory to clone repos into")
	rootCmd.AddCommand(runCmd, checkCmd, tailscaleCmd, toolsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func step(msg string) {
	fmt.Printf("\n▶  %s\n", msg)
}

func banner(msg string) {
	line := "══════════════════════════════════════════════"
	fmt.Printf("\n╔%s╗\n║  %-44s║\n╚%s╝\n", line, msg, line)
}
