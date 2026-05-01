// internal/cmd/sessions.go
// "sessions" subcommand — lists all saved sessions.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hydria-ai/hydria/internal/session"
	"github.com/hydria-ai/hydria/internal/tracker"
	"github.com/hydria-ai/hydria/internal/ui"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List all saved sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := tracker.Init(dbPath()); err != nil {
			return fmt.Errorf("init db: %w", err)
		}

		sessions, err := session.List()
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}

		rows := make([]ui.SessionRow, len(sessions))
		for i, s := range sessions {
			rows[i] = ui.SessionRow{
				SessionID:     s.SessionID,
				Target:        s.Target,
				Service:       s.Service,
				Status:        s.Status,
				CreatedAt:     s.CreatedAt,
				FoundPassword: s.FoundPassword,
			}
		}
		ui.ListSessionsTable(rows)
		return nil
	},
}
