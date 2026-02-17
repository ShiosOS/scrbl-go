package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/juliuswalton/scrbl/internal/config"
	"github.com/juliuswalton/scrbl/internal/dayfiles"
	syncclient "github.com/juliuswalton/scrbl/sync"
)

func runSync(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing sync subcommand (supported: push, pull)")
	}

	switch args[0] {
	case "push":
		return runSyncPush(args[1:])
	case "pull":
		return runSyncPull(args[1:])
	default:
		return fmt.Errorf("unknown sync subcommand %q", args[0])
	}
}

func runSyncPush(args []string) error {
	fs := flag.NewFlagSet("sync push", flag.ContinueOnError)
	date := fs.String("date", "", "date to sync (YYYY-MM-DD), default today")
	all := fs.Bool("all", false, "push all local day files")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("sync push does not take positional arguments")
	}
	if *all && strings.TrimSpace(*date) != "" {
		return fmt.Errorf("--all and --date cannot be used together")
	}

	cfg, client, err := loadSyncClient()
	if err != nil {
		return err
	}

	if *all {
		return pushAll(cfg, client)
	}

	day, err := dayfiles.ParseDateOrToday(*date)
	if err != nil {
		return err
	}

	content, err := dayfiles.Read(cfg.NotesDir, day)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("local note not found for %s", day.Format(dayfiles.DateLayout))
		}
		return err
	}

	if err := client.PushNote(day, content); err != nil {
		return err
	}

	fmt.Printf("pushed %s\n", day.Format(dayfiles.DateLayout))
	return nil
}

func runSyncPull(args []string) error {
	fs := flag.NewFlagSet("sync pull", flag.ContinueOnError)
	date := fs.String("date", "", "date to sync (YYYY-MM-DD), default today")
	all := fs.Bool("all", false, "pull all remote days")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("sync pull does not take positional arguments")
	}
	if *all && strings.TrimSpace(*date) != "" {
		return fmt.Errorf("--all and --date cannot be used together")
	}

	cfg, client, err := loadSyncClient()
	if err != nil {
		return err
	}

	if *all {
		return pullAll(cfg, client)
	}

	day, err := dayfiles.ParseDateOrToday(*date)
	if err != nil {
		return err
	}

	content, err := client.PullNote(day)
	if err != nil {
		return err
	}
	if content == "" {
		return fmt.Errorf("remote note not found for %s", day.Format(dayfiles.DateLayout))
	}

	if err := dayfiles.Write(cfg.NotesDir, day, content); err != nil {
		return err
	}

	fmt.Printf("pulled %s\n", day.Format(dayfiles.DateLayout))
	return nil
}

func loadSyncClient() (config.Config, *syncclient.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, nil, err
	}

	client := syncclient.NewClient(cfg.ServerURL, cfg.APIKey)
	if client == nil {
		return config.Config{}, nil, fmt.Errorf("server_url is not configured (run: scrbl init --server <url>)")
	}

	if err := os.MkdirAll(cfg.NotesDir, 0o755); err != nil {
		return config.Config{}, nil, fmt.Errorf("create notes dir: %w", err)
	}

	return cfg, client, nil
}

func pushAll(cfg config.Config, client *syncclient.Client) error {
	dates, err := dayfiles.ListDates(cfg.NotesDir)
	if err != nil {
		return err
	}
	if len(dates) == 0 {
		fmt.Println("no local day files to push")
		return nil
	}

	pushed := 0
	failed := 0

	for _, day := range dates {
		content, err := dayfiles.Read(cfg.NotesDir, day)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", day.Format(dayfiles.DateLayout), err)
			failed++
			continue
		}

		if err := client.PushNote(day, content); err != nil {
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", day.Format(dayfiles.DateLayout), err)
			failed++
			continue
		}

		fmt.Printf("pushed %s\n", day.Format(dayfiles.DateLayout))
		pushed++
	}

	fmt.Printf("push complete: %d pushed, %d failed\n", pushed, failed)
	if failed > 0 {
		return fmt.Errorf("push completed with failures")
	}

	return nil
}

func pullAll(cfg config.Config, client *syncclient.Client) error {
	dateStrings, err := client.PullAllDates()
	if err != nil {
		return err
	}
	if len(dateStrings) == 0 {
		fmt.Println("no remote day files to pull")
		return nil
	}

	sort.Strings(dateStrings)
	pulled := 0
	skipped := 0
	failed := 0

	for _, ds := range dateStrings {
		day, err := time.Parse(dayfiles.DateLayout, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: invalid date from server\n", ds)
			skipped++
			continue
		}

		content, err := client.PullNote(day)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", ds, err)
			failed++
			continue
		}
		if content == "" {
			skipped++
			continue
		}

		if err := dayfiles.Write(cfg.NotesDir, day, content); err != nil {
			fmt.Fprintf(os.Stderr, "fail %s: %v\n", ds, err)
			failed++
			continue
		}

		fmt.Printf("pulled %s\n", ds)
		pulled++
	}

	fmt.Printf("pull complete: %d pulled, %d skipped, %d failed\n", pulled, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("pull completed with failures")
	}

	return nil
}
