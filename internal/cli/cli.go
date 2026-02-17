package cli

import (
	"fmt"
)

func Run(args []string) error {
	if len(args) == 0 {
		return runTUI(nil)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "init":
		return runInit(args[1:])
	case "config":
		return runConfig(args[1:])
	case "tui":
		return runTUI(args[1:])
	case "migrate":
		return runMigrate(args[1:])
	case "summary":
		return runSummary(args[1:])
	case "sync":
		return runSync(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("scrbl")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                Create/update local CLI config")
	fmt.Println("  config show         Print current config")
	fmt.Println("  tui                 Open notes stream + embedded neovim composer")
	fmt.Println("  migrate             Migrate local note format")
	fmt.Println("  summary             Copy latest ## Summary as Slack markdown")
	fmt.Println("  sync push           Push local note(s) to the server")
	fmt.Println("  sync pull           Pull remote note(s) into local notes")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scrbl init --server https://scrbl.example.com --api-key <key>")
	fmt.Println("  scrbl tui")
	fmt.Println("  scrbl summary")
	fmt.Println("  scrbl migrate --sync")
	fmt.Println("  scrbl sync push --date 2026-02-17")
	fmt.Println("  scrbl sync push --all")
	fmt.Println("  scrbl sync pull --all")
}
