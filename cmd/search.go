package cmd

import (
	"fmt"
	"strings"

	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search across all notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := notes.NewStore(cfg.NotesDir)
		query := strings.Join(args, " ")

		results, err := store.Search(query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("  No results found.")
			return nil
		}

		fmt.Printf("  Found %d results for \"%s\":\n\n", len(results), query)
		for _, r := range results {
			dateStr := r.Date.Format("Jan 2, 2006")
			fmt.Printf("  %s  %s\n", dateStr, r.Line)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
