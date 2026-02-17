package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/juliuswalton/scrbl/internal/config"
	"github.com/juliuswalton/scrbl/notes"
	syncclient "github.com/juliuswalton/scrbl/sync"
	"github.com/juliuswalton/scrbl/tui"
)

func runTUI(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	editor := fs.String("editor", "", "editor binary for embedded vim/neovim")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("tui does not take positional arguments")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.NotesDir, 0o755); err != nil {
		return fmt.Errorf("create notes dir: %w", err)
	}

	ed := strings.TrimSpace(*editor)
	if ed == "" {
		ed = "nvim"
	}

	store := notes.NewStore(cfg.NotesDir)
	syncer := syncclient.NewClient(cfg.ServerURL, cfg.APIKey)
	app := tui.NewApp(store, syncer, ed)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	return nil
}
