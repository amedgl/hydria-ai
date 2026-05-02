// internal/cmd/root.go
// Root cobra command, shared flags, and shared helpers (data dir, db path).
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hydria-ai/hydria/internal/ui"
)

// Flags holds all CLI flag values, shared across attack.go and sessions.go.
var flags struct {
	Image     string
	Text      string // keyword text input (alternative or complement to --image)
	Target    string
	Service   string
	Username  string
	Port      int
	APIKey    string
	Model     string
	SessionID string
	Threads   int
	BatchSize int
	DryRun    bool
}

// RootCmd is the top-level cobra command.
var RootCmd = &cobra.Command{
	Use:   "hydria",
	Short: "HydrIA AI — Gemini Vision powered THC-Hydra attack framework",
	Long: `HydrIA AI — AI-Powered Attack Framework

Analyze a target image OR keyword text with Gemini → Generate a smart wordlist → Attack with THC-Hydra.
You can also combine --image and --text to merge both analysis results.

⚠  For authorized systems only. Unauthorized access is illegal.`,
	Example: `  hydria -i target.jpg -t 192.168.1.10 -s ssh -u admin
  hydria --text "john doe 1990 istanbul fenerbahce" -t 192.168.1.10 -s ssh -u admin
  hydria -i target.jpg --text "extra hints" -t 192.168.1.10 -s ssh -u admin --dry-run`,
	RunE: runAttack,
}

func init() {
	f := RootCmd.Flags()
	f.StringVarP(&flags.Image, "image", "i", "", "Path to target image file")
	f.StringVar(&flags.Text, "text", "", "Keyword hints about the target (name, date, city, hobby, ...)")
	f.StringVarP(&flags.Target, "target", "t", "", "Target IP or domain")
	f.StringVarP(&flags.Service, "service", "s", "", "Protocol (ssh, ftp, rdp, ...)")
	f.StringVarP(&flags.Username, "username", "u", "", "Username")
	f.IntVarP(&flags.Port, "port", "p", 0, "Port (optional)")
	f.StringVar(&flags.APIKey, "api-key", "", "Gemini API key (.env default)")
	f.StringVar(&flags.Model, "model", "", "Gemini model (default: gemini-flash-latest)")
	f.StringVar(&flags.SessionID, "session", "", "Session ID to resume")
	f.IntVar(&flags.Threads, "threads", 0, "Parallel threads (default: 4)")
	f.IntVar(&flags.BatchSize, "batch-size", 0, "Passwords per batch (default: 50)")
	f.BoolVar(&flags.DryRun, "dry-run", false, "Generate wordlist only, do not run Hydra")

	RootCmd.AddCommand(sessionsCmd)
}

// Execute is the single entry point called from main.go.
func Execute() {
	ui.PrintBanner()
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── Shared helpers ────────────────────────────────────────────────────────

func dataDir() string {
	dir := filepath.Join(".", "data")
	os.MkdirAll(dir, 0755) //nolint:errcheck
	return dir
}

func dbPath() string {
	return filepath.Join(dataDir(), "hydria.db")
}
