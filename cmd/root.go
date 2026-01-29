package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/juliuswalton/scrbl/sync"
	"github.com/juliuswalton/scrbl/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scrbl [message]",
	Short: "A zero-organization note stream for your terminal",
	Long:  "scrbl is an append-only, chronological note stream. Just write.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store := notes.NewStore(cfg.NotesDir)
		if err := store.EnsureDir(); err != nil {
			return fmt.Errorf("failed to create notes dir: %w", err)
		}

		syncer := sync.NewClient(cfg.ServerURL, cfg.APIKey)

		// If args provided, quick-append mode
		if len(args) > 0 {
			content := strings.Join(args, " ")
			today := notes.Today()
			if err := store.AppendEntry(today, content, false); err != nil {
				return fmt.Errorf("failed to append note: %w", err)
			}

			// Sync silently
			if syncer != nil {
				dayContent, _ := store.ReadDay(today)
				_ = syncer.PushNote(today, dayContent)
			}

			fmt.Println("  noted.")
			return nil
		}

		// Launch TUI
		app := tui.NewApp(store, syncer, cfg.Editor)
		p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
