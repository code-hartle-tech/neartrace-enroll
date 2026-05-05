package repos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Repo struct {
	Name string
	URL  string // HTTPS, token injected at clone time
}

// HarlteTechRepos is the canonical list of repos every team member needs.
var HarlteTechRepos = []Repo{
	{Name: "neartrace-android-mvp", URL: "https://github.com/code-hartle-tech/neartrace-android-mvp.git"},
	{Name: "neartrace-docs", URL: "https://github.com/code-hartle-tech/neartrace-docs.git"},
	{Name: "neartrace-web", URL: "https://github.com/code-hartle-tech/neartrace-web.git"},
	{Name: "hartle.tech-terraform", URL: "https://github.com/code-hartle-tech/hartle.tech-terraform.git"},
	{Name: "neartrace-enroll", URL: "https://github.com/code-hartle-tech/neartrace-enroll.git"},
}

// CloneAll clones each repo into baseDir, skipping any that already exist.
// token is a GitHub PAT injected into the HTTPS URL.
func CloneAll(baseDir, token string) error {
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return fmt.Errorf("create base dir: %w", err)
	}
	for _, r := range HarlteTechRepos {
		dest := filepath.Join(baseDir, r.Name)
		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("  ✅  %-35s already present\n", r.Name)
			continue
		}
		fmt.Printf("  →   %-35s cloning...\n", r.Name)
		url := injectToken(r.URL, token)
		cmd := exec.Command("git", "clone", "--quiet", url, dest)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("clone %s: %w\n%s", r.Name, err, out)
		}
		fmt.Printf("  ✅  %-35s cloned\n", r.Name)
	}
	return nil
}

func injectToken(httpsURL, token string) string {
	if token == "" {
		return httpsURL
	}
	// https://github.com/... → https://oauth2:<token>@github.com/...
	const prefix = "https://"
	if len(httpsURL) > len(prefix) {
		return prefix + "oauth2:" + token + "@" + httpsURL[len(prefix):]
	}
	return httpsURL
}
