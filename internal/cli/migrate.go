package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/juliuswalton/scrbl/internal/config"
	"github.com/juliuswalton/scrbl/internal/dayfiles"
	"github.com/juliuswalton/scrbl/internal/migrate"
)

func runMigrate(args []string) error {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	syncAfter := fs.Bool("sync", false, "push all notes to remote server after migration")
	dryRun := fs.Bool("dry-run", false, "show what would change without writing files")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("migrate does not take positional arguments")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.NotesDir, 0o755); err != nil {
		return fmt.Errorf("create notes dir: %w", err)
	}

	dates, err := dayfiles.ListDates(cfg.NotesDir)
	if err != nil {
		return err
	}
	if len(dates) == 0 {
		fmt.Println("no local notes to migrate")
		return nil
	}

	checked := 0
	changedCount := 0

	for _, day := range dates {
		checked++

		raw, err := dayfiles.Read(cfg.NotesDir, day)
		if err != nil {
			return err
		}

		updated, changed := migrate.NormalizeDayContent(day, raw)
		if !changed {
			continue
		}

		if *dryRun {
			fmt.Printf("would migrate %s\n", day.Format(dayfiles.DateLayout))
			changedCount++
			continue
		}

		if err := dayfiles.Write(cfg.NotesDir, day, updated); err != nil {
			return err
		}

		fmt.Printf("migrated %s\n", day.Format(dayfiles.DateLayout))
		changedCount++
	}

	if *dryRun {
		fmt.Printf("dry-run complete: %d checked, %d would change\n", checked, changedCount)
	} else {
		fmt.Printf("migration complete: %d checked, %d changed\n", checked, changedCount)
	}

	if !*syncAfter {
		return nil
	}
	if *dryRun {
		fmt.Println("dry-run mode: skipping sync")
		return nil
	}

	cfg, client, err := loadSyncClient()
	if err != nil {
		return err
	}

	return pushAll(cfg, client)
}
