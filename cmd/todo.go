package cmd

import (
	"fmt"
	"strings"

	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/juliuswalton/scrbl/sync"
	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo [task description]",
	Short: "Add a task to today's notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := notes.NewStore(cfg.NotesDir)
		if err := store.EnsureDir(); err != nil {
			return err
		}

		content := strings.Join(args, " ")
		today := notes.Today()

		if err := store.AppendEntry(today, content, true); err != nil {
			return fmt.Errorf("failed to add task: %w", err)
		}

		// Sync silently
		syncer := sync.NewClient(cfg.ServerURL, cfg.APIKey)
		if syncer != nil {
			dayContent, _ := store.ReadDay(today)
			_ = syncer.PushNote(today, dayContent)
		}

		fmt.Printf("  task added: %s\n", content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todoCmd)
}
