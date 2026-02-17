package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/juliuswalton/scrbl/internal/config"
)

func runInit(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	notesDir := fs.String("notes-dir", cfg.NotesDir, "directory where local day markdown files live")
	serverURL := fs.String("server", cfg.ServerURL, "sync server URL")
	apiKey := fs.String("api-key", cfg.APIKey, "sync API key")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("init does not take positional arguments")
	}

	cfg.NotesDir = strings.TrimSpace(*notesDir)
	cfg.ServerURL = strings.TrimSpace(*serverURL)
	cfg.APIKey = strings.TrimSpace(*apiKey)

	if err := config.Save(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.NotesDir, 0o755); err != nil {
		return fmt.Errorf("create notes dir: %w", err)
	}

	fmt.Println("saved config:")
	fmt.Println("  file:", config.Path())
	fmt.Println("  notes_dir:", cfg.NotesDir)
	fmt.Println("  server_url:", cfg.ServerURL)
	if cfg.APIKey != "" {
		fmt.Println("  api_key: [set]")
	} else {
		fmt.Println("  api_key: [empty]")
	}

	return nil
}

func runConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing config subcommand (supported: show)")
	}

	switch args[0] {
	case "show":
		if len(args) > 1 {
			return fmt.Errorf("config show does not take positional arguments")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	default:
		return fmt.Errorf("unknown config subcommand %q", args[0])
	}
}
