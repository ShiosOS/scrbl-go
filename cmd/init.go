package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/juliuswalton/scrbl/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up scrbl configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println()
		fmt.Println("  Welcome to scrbl.")
		fmt.Println("  Let's get you set up.\n")

		// Server URL
		fmt.Print("  Server URL (leave blank to skip sync): ")
		serverURL, _ := reader.ReadString('\n')
		serverURL = strings.TrimSpace(serverURL)

		// API Key
		var apiKey string
		if serverURL != "" {
			fmt.Print("  API Key: ")
			apiKey, _ = reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKey)
		}

		// Editor
		fmt.Print("  Editor [nvim]: ")
		editor, _ := reader.ReadString('\n')
		editor = strings.TrimSpace(editor)
		if editor == "" {
			editor = "nvim"
		}

		cfg := &config.Config{
			NotesDir:  config.DefaultNotesDir(),
			ServerURL: serverURL,
			APIKey:    apiKey,
			Editor:    editor,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Printf("  Config saved to %s/config.yaml\n", config.DefaultDir())
		fmt.Printf("  Notes will be stored in %s\n", cfg.NotesDir)
		if serverURL != "" {
			fmt.Printf("  Syncing to %s\n", serverURL)
		} else {
			fmt.Println("  Sync disabled (run 'scrbl init' again to set up)")
		}
		fmt.Println("\n  You're good to go. Run 'scrbl' to start writing.")
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
