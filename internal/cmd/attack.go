// internal/cmd/attack.go
// Core attack orchestration — image analysis, wordlist generation, Hydra execution.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/hydria-ai/hydria/internal/config"
	"github.com/hydria-ai/hydria/internal/hydra"
	"github.com/hydria-ai/hydria/internal/session"
	"github.com/hydria-ai/hydria/internal/tracker"
	"github.com/hydria-ai/hydria/internal/ui"
	"github.com/hydria-ai/hydria/internal/vision"
	"github.com/hydria-ai/hydria/internal/wordlist"
)

// runAttack is the RunE handler for the root command.
func runAttack(cmd *cobra.Command, args []string) error {
	cfg := config.Load()

	// ── API key ────────────────────────────────────────────────────────
	apiKey := flags.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		ui.PrintError("Gemini API key not found!")
		ui.PrintInfo("  • Add GEMINI_API_KEY=your_key to the .env file")
		ui.PrintInfo("  • or pass it via --api-key")
		os.Exit(1)
	}

	// ── Required flags ─────────────────────────────────────────────────
	if flags.Target == "" || flags.Service == "" || flags.Username == "" {
		ui.PrintError("--target, --service, and --username are required.")
		os.Exit(1)
	}

	// ── Hydra presence check ───────────────────────────────────────────
	if !flags.DryRun && !hydra.IsInstalled() {
		ui.PrintError("THC-Hydra is not installed!")
		ui.PrintInfo("  Install it with: sudo apt install hydra")
		os.Exit(1)
	}

	// ── Database ───────────────────────────────────────────────────────
	if err := tracker.Init(dbPath()); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	// ── Session management ─────────────────────────────────────────────
	sessionID, resumed, existingWordlist, err := resolveSession(cfg)
	if err != nil {
		return err
	}

	ui.PrintSessionInfo(sessionID, flags.Target, flags.Service, flags.Username, resumed)

	// ── Image analysis + wordlist generation ───────────────────────────
	passwords, err := buildPasswords(cfg, sessionID, resumed, existingWordlist, apiKey)
	if err != nil {
		return err
	}

	if flags.DryRun {
		ui.PrintSuccess("--dry-run mode: wordlist generated, Hydra not launched.")
		session.UpdateStatus(sessionID, "completed", "")
		return nil
	}

	// ── Attack ─────────────────────────────────────────────────────────
	return launchAttack(cfg, sessionID, passwords)
}

// resolveSession creates a new session or resumes a paused one.
func resolveSession(cfg config.Config) (sessionID string, resumed bool, existingWordlist string, err error) {
	if flags.SessionID != "" {
		s, e := session.Get(flags.SessionID)
		if e != nil || s == nil {
			ui.PrintError(fmt.Sprintf("Session not found: %s", flags.SessionID))
			os.Exit(1)
		}
		session.UpdateStatus(flags.SessionID, "running", "")
		return flags.SessionID, true, s.WordlistPath, nil
	}

	if cfg.Session.AutoResume {
		paused, _ := session.FindResumable(flags.Target, flags.Username)
		if paused != nil {
			fmt.Printf("\n  ⟳  Paused session found: %s\n", paused.SessionID)
			fmt.Print("  Press Enter to resume, or type 'n' to start fresh: ")
			reader := bufio.NewReader(os.Stdin)
			choice, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(choice)) != "n" {
				session.UpdateStatus(paused.SessionID, "running", "")
				return paused.SessionID, true, paused.WordlistPath, nil
			}
		}
	}

	id, e := session.Create(flags.Target, flags.Service, flags.Username, flags.Image)
	if e != nil {
		return "", false, "", fmt.Errorf("create session: %w", e)
	}
	return id, false, "", nil
}

// buildPasswords handles analysis and wordlist generation (or loads existing).
// Supports three input modes:
//   --image only   → Gemini Vision analysis
//   --text only    → Gemini text/keyword analysis
//   --image + text → both analyses merged into one result
func buildPasswords(cfg config.Config, sessionID string, resumed bool, existingWordlist, apiKey string) ([]string, error) {
	if resumed && existingWordlist != "" {
		ui.PrintInfo(fmt.Sprintf("Using existing wordlist: %s", existingWordlist))
		passwords, err := wordlist.Load(existingWordlist)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Load wordlist: %v", err))
			os.Exit(1)
		}
		ui.PrintInfo(fmt.Sprintf("%d passwords loaded", len(passwords)))
		return passwords, nil
	}

	// Must have at least one input source
	if flags.Image == "" && strings.TrimSpace(flags.Text) == "" {
		ui.PrintError("Provide at least one input: --image <file>  or  --text \"keywords\"")
		os.Exit(1)
	}

	modelName := flags.Model
	if modelName == "" {
		modelName = cfg.Gemini.Model
	}

	ctx := context.Background()
	var finalAnalysis vision.AnalysisResult

	// ── Image analysis ─────────────────────────────────────────────────
	if flags.Image != "" {
		ui.PrintSection("🧠", "Gemini Vision Analysis")
		ui.PrintInfo(fmt.Sprintf("Loading image: %s", flags.Image))

		imgAnalysis, err := vision.AnalyzeImage(ctx, flags.Image, apiKey, modelName)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Image analysis error: %v", err))
			os.Exit(1)
		}
		ui.PrintSuccess(fmt.Sprintf("%d hints from image", imgAnalysis.CountHints()))
		finalAnalysis = vision.MergeResults(finalAnalysis, imgAnalysis)
	}

	// ── Text / keyword analysis ────────────────────────────────────────
	if strings.TrimSpace(flags.Text) != "" {
		ui.PrintSection("💬", "Gemini Text Analysis")
		ui.PrintInfo(fmt.Sprintf("Keywords: %s", flags.Text))

		txtAnalysis, err := vision.AnalyzeText(ctx, flags.Text, apiKey, modelName)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Text analysis error: %v", err))
			os.Exit(1)
		}
		ui.PrintSuccess(fmt.Sprintf("%d hints from keywords", txtAnalysis.CountHints()))
		finalAnalysis = vision.MergeResults(finalAnalysis, txtAnalysis)
	}

	// ── Show combined results ──────────────────────────────────────────
	ui.PrintSuccess(fmt.Sprintf("%d total hints (combined)", finalAnalysis.CountHints()))
	rows := finalAnalysis.ToDisplayRows()
	displayRows := make([]ui.AnalysisDisplay, len(rows))
	for i, r := range rows {
		displayRows[i] = ui.AnalysisDisplay{
			Category: r.Category,
			Values:   strings.Split(r.Values, ", "),
		}
	}
	ui.PrintAnalysisResults(displayRows)

	// ── Wordlist generation ────────────────────────────────────────────
	ui.PrintSection("📋", "Wordlist Generation")

	passwords := wordlist.Generate(finalAnalysis, wordlist.Options{
		MinLength:           cfg.Wordlist.MinLength,
		MaxLength:           cfg.Wordlist.MaxLength,
		MaxSize:             cfg.Wordlist.MaxSize,
		IncludeLeet:         cfg.Wordlist.IncludeLeet,
		IncludeReverse:      cfg.Wordlist.IncludeReverse,
		IncludeCombinations: cfg.Wordlist.IncludeCombinations,
	})

	wlPath := wordlist.Filename(dataDir(), flags.Target, sessionID)
	if err := wordlist.Save(passwords, wlPath); err != nil {
		ui.PrintError(fmt.Sprintf("Save wordlist: %v", err))
		os.Exit(1)
	}
	session.UpdateWordlist(sessionID, wlPath)
	ui.PrintWordlistStats(len(passwords), wlPath)

	return passwords, nil
}

// launchAttack runs Hydra and handles the result + Ctrl+C.
func launchAttack(cfg config.Config, sessionID string, passwords []string) error {
	ui.PrintSection("⚔", "Launching THC-Hydra Attack")

	alreadyTried := tracker.GetTriedCount(flags.Target, flags.Username)
	if alreadyTried > 0 {
		ui.PrintInfo(fmt.Sprintf("%d passwords already tried — will be skipped", alreadyTried))
	}

	threads := flags.Threads
	if threads == 0 {
		threads = cfg.Hydra.Threads
	}
	batchSize := flags.BatchSize
	if batchSize == 0 {
		batchSize = cfg.Hydra.BatchSize
	}

	bar := progressbar.NewOptions(len(passwords),
		progressbar.OptionSetDescription("Attacking "+flags.Target),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer: "=", SaucerPadding: " ", BarStart: "[", BarEnd: "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	startTime := time.Now()
	globalTried := 0

	resultCh := make(chan hydra.Result, 1)
	errCh := make(chan error, 1)

	go func() {
		res, err := hydra.RunAttack(
			flags.Target, flags.Service, flags.Username,
			passwords, flags.Port, threads, batchSize,
			func(tried, total int, current string) {
				globalTried = tried
				bar.Set(tried) //nolint:errcheck
			},
		)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	select {
	case <-sigCh:
		fmt.Println()
		ui.PrintWarning("Attack paused.")
		session.Pause(sessionID)
		session.UpdateCounts(sessionID, globalTried, 0)
		ui.PrintInfo(fmt.Sprintf(
			"  To resume: hydria --session %s -t %s -s %s -u %s",
			sessionID, flags.Target, flags.Service, flags.Username,
		))
		return nil

	case err := <-errCh:
		return fmt.Errorf("attack error: %w", err)

	case res := <-resultCh:
		bar.Finish() //nolint:errcheck
		elapsed := time.Since(startTime)
		session.UpdateCounts(sessionID, res.Tried, res.Skipped)

		if res.Found {
			ui.PrintFound(res.Password, flags.Username, flags.Target)
			session.UpdateStatus(sessionID, "completed", res.Password)
		} else {
			fmt.Println()
			ui.PrintWarning("Wordlist exhausted — password not found.")
			ui.PrintInfo(fmt.Sprintf("  Tried   : %d passwords", res.Tried))
			ui.PrintInfo(fmt.Sprintf("  Skipped : %d (already tried)", res.Skipped))
			ui.PrintInfo(fmt.Sprintf("  Duration: %dm %ds",
				int(elapsed.Minutes()), int(elapsed.Seconds())%60))
			session.UpdateStatus(sessionID, "completed", "")
		}
	}

	return nil
}
