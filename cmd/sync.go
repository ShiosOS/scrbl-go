package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all local notes to the remote server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		syncer := sync.NewClient(cfg.ServerURL, cfg.APIKey)
		if syncer == nil {
			return fmt.Errorf("no server_url configured in ~/.scrbl/config.yaml")
		}

		// List all local day files
		entries, err := os.ReadDir(cfg.NotesDir)
		if err != nil {
			return fmt.Errorf("failed to read notes dir: %w", err)
		}

		var pushed, failed int
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}

			date := strings.TrimSuffix(name, ".md")
			content, err := os.ReadFile(filepath.Join(cfg.NotesDir, name))
			if err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL %s: %v\n", date, err)
				failed++
				continue
			}

			parsed, err := time.Parse("2006-01-02", date)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  SKIP %s: invalid date\n", name)
				continue
			}

			if err := syncer.PushNote(parsed, string(content)); err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL %s: %v\n", date, err)
				failed++
				continue
			}

			fmt.Printf("  OK   %s\n", date)
			pushed++
		}

		fmt.Printf("\nDone: %d pushed, %d failed\n", pushed, failed)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
