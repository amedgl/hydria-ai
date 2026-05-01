// internal/ui/ui.go
// Rich terminal UI for HydrIA AI — banner, panels, progress, tables.
package ui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────

var (
	boldRed    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4444"))
	boldGreen  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF88"))
	boldCyan   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00CCFF"))
	boldYellow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
	boldBlue   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4488FF"))
	boldWhite  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dim        = lipgloss.NewStyle().Faint(true)
	cyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CCFF"))
	green      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))
	yellow     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	magenta    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF88FF"))
)

// ── Banner ────────────────────────────────────────────────────────────────

const bannerArt = `
  ██╗  ██╗██╗   ██╗██████╗ ██████╗ ██╗ █████╗      █████╗ ██╗
  ██║  ██║╚██╗ ██╔╝██╔══██╗██╔══██╗██║██╔══██╗    ██╔══██╗██║
  ███████║ ╚████╔╝ ██║  ██║██████╔╝██║███████║    ███████║██║
  ██╔══██║  ╚██╔╝  ██║  ██║██╔══██╗██║██╔══██║    ██╔══██║██║
  ██║  ██║   ██║   ██████╔╝██║  ██║██║██║  ██║    ██║  ██║██║
  ╚═╝  ╚═╝   ╚═╝   ╚═════╝ ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝   ╚═╝  ╚═╝╚═╝`

// PrintBanner prints the HydrIA AI ASCII banner.
func PrintBanner() {
	fmt.Println(boldRed.Render(bannerArt))
	fmt.Println(dim.Render("                 AI-Powered Attack Framework"))
	fmt.Println(boldYellow.Render("            ⚠  For authorized systems only."))
	fmt.Println()
}

// ── Section / status helpers ──────────────────────────────────────────────

func PrintSection(emoji, title string) {
	fmt.Printf("\n%s  %s\n", boldCyan.Render(emoji), boldCyan.Render(title))
	fmt.Println(dim.Render(strings.Repeat("─", 50)))
}

func PrintSuccess(msg string) {
	fmt.Printf("  %s  %s\n", boldGreen.Render("✓"), msg)
}

func PrintInfo(msg string) {
	fmt.Printf("  %s  %s\n", boldBlue.Render("ℹ"), msg)
}

func PrintWarning(msg string) {
	fmt.Printf("  %s  %s\n", boldYellow.Render("⚠"), msg)
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "  %s  %s\n", boldRed.Render("✗"), msg)
}

// ── Panels ────────────────────────────────────────────────────────────────

var panelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 2)

func PrintFound(password, username, target string) {
	fmt.Println()
	content := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s",
		boldWhite.Render("Target  :"), cyan.Render(target),
		boldWhite.Render("Username:"), cyan.Render(username),
		boldWhite.Render("Password:"), boldGreen.Render(password),
	)
	panel := panelStyle.
		BorderForeground(lipgloss.Color("#00FF88")).
		Render(content)
	fmt.Printf("\n  %s\n%s\n", boldGreen.Render("🔓  PASSWORD FOUND!"), panel)
}

func PrintWordlistStats(total int, path string) {
	content := fmt.Sprintf(
		"%s %s\n%s %s",
		boldWhite.Render("Total Passwords :"), boldGreen.Render(fmt.Sprintf("%d", total)),
		boldWhite.Render("File            :"), dim.Render(path),
	)
	panel := panelStyle.
		BorderForeground(lipgloss.Color("#4488FF")).
		Render(content)
	fmt.Printf("\n  %s\n%s\n", boldCyan.Render("📋  Wordlist Ready"), panel)
}

func PrintSessionInfo(sessionID, target, service, username string, resumed bool) {
	statusStr := green.Render("New Session")
	if resumed {
		statusStr = yellow.Render("Resumed")
	}
	content := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
		boldWhite.Render("Session  :"), dim.Render(sessionID),
		boldWhite.Render("Target   :"), cyan.Render(target),
		boldWhite.Render("Service  :"), cyan.Render(service),
		boldWhite.Render("Username :"), cyan.Render(username),
		boldWhite.Render("Status   :"), statusStr,
	)
	panel := panelStyle.
		BorderForeground(lipgloss.Color("#FFD700")).
		Render(content)
	fmt.Printf("\n  %s\n%s\n", boldYellow.Render("🎯  Attack Session"), panel)
}

// ── Analysis results table ────────────────────────────────────────────────

type AnalysisDisplay struct {
	Category string
	Values   []string
}

func PrintAnalysisResults(rows []AnalysisDisplay) {
	fmt.Printf("\n  %s\n", magenta.Render("🧠  Gemini Vision Analysis Results"))
	fmt.Println(dim.Render("  " + strings.Repeat("─", 52)))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, magenta.Render("  Category")+"\t"+magenta.Render("Found Values"))
	fmt.Fprintln(w, dim.Render("  --------")+"\t"+dim.Render("------------"))
	for _, row := range rows {
		if len(row.Values) > 0 {
			fmt.Fprintf(w, "  %s\t%s\n",
				boldWhite.Render(row.Category),
				strings.Join(row.Values, ", "),
			)
		}
	}
	w.Flush()
	fmt.Println()
}

// ── Sessions list table ───────────────────────────────────────────────────

type SessionRow struct {
	SessionID     string
	Target        string
	Service       string
	Status        string
	CreatedAt     string
	FoundPassword string
}

func ListSessionsTable(sessions []SessionRow) {
	if len(sessions) == 0 {
		PrintWarning("No saved sessions found.")
		return
	}

	fmt.Printf("\n  %s\n", magenta.Render("📂  Saved Sessions"))
	fmt.Println(dim.Render("  " + strings.Repeat("─", 80)))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\n",
		magenta.Render("Session ID"),
		magenta.Render("Target"),
		magenta.Render("Service"),
		magenta.Render("Status"),
		magenta.Render("Created At"),
		magenta.Render("Found Password"),
	)
	for _, s := range sessions {
		var statusStyled string
		switch s.Status {
		case "running":
			statusStyled = yellow.Render(s.Status)
		case "completed":
			statusStyled = green.Render(s.Status)
		case "paused":
			statusStyled = boldRed.Render(s.Status)
		default:
			statusStyled = s.Status
		}
		found := s.FoundPassword
		if found == "" {
			found = dim.Render("—")
		} else {
			found = boldGreen.Render(found)
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\n",
			dim.Render(s.SessionID),
			cyan.Render(s.Target),
			s.Service,
			statusStyled,
			dim.Render(s.CreatedAt),
			found,
		)
	}
	w.Flush()
}
