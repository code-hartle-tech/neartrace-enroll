package tokens

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirMode  = 0700
	FileMode = 0600
)

var tokenStub = `# HARTLE.TECH infrastructure credentials
# chmod 600 — never commit, never share, never paste into chat
# source before infra work: source ~/.config/hartle.tech/tokens

# ── Cloudflare ────────────────────────────────────────────────────────────────
export CF_API_TOKEN=""
export CF_ACCOUNT_ID="38e54e7b46f5f224dc37de3cceebfd7f"
export CF_ZONE_ID_NEARTRACE="a36fd6cb0871301768018f64fe953c9a"

# ── GitHub ────────────────────────────────────────────────────────────────────
export GITHUB_TOKEN=""

# ── Tailscale OAuth (for Terraform) ──────────────────────────────────────────
# Create at: tailscale.com/admin/settings/oauth
export TAILSCALE_OAUTH_CLIENT_ID=""
export TAILSCALE_OAUTH_CLIENT_SECRET=""

# ── Terraform convenience exports ─────────────────────────────────────────────
export TF_VAR_cloudflare_api_token="$CF_API_TOKEN"
export TF_VAR_tailscale_oauth_client_id="$TAILSCALE_OAUTH_CLIENT_ID"
export TF_VAR_tailscale_oauth_client_secret="$TAILSCALE_OAUTH_CLIENT_SECRET"
`

// Dir returns the platform-appropriate config directory.
func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hartle.tech")
}

// File returns the full path to the tokens file.
func File() string {
	return filepath.Join(Dir(), "tokens")
}

// Seed creates the tokens file with a commented stub if it does not already exist.
// Never overwrites an existing file.
func Seed() error {
	dir := Dir()
	if err := os.MkdirAll(dir, DirMode); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := File()
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  ✅  %s already exists — not overwritten\n", path)
		return nil
	}

	if err := os.WriteFile(path, []byte(tokenStub), FileMode); err != nil {
		return fmt.Errorf("write tokens file: %w", err)
	}
	fmt.Printf("  ✅  Created %s\n", path)
	fmt.Printf("  ⚠️   Fill in the empty values before running terraform\n")
	return nil
}
