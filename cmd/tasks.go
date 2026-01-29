package cmd

import (
	"fmt"

	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List all open tasks across all notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := notes.NewStore(cfg.NotesDir)

		tasks, err := store.GetOpenTasks()
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("  No open tasks. Nice work.")
			return nil
		}

		fmt.Printf("  %d open tasks:\n\n", len(tasks))
		for _, t := range tasks {
			dateStr := t.Date.Format("Jan 2")
			fmt.Printf("  %s  %s\n", dateStr, t.Line)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}
