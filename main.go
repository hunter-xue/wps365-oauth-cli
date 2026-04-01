package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"oauth_tools/cmd"
	"oauth_tools/config"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `oauth_tools — fetch OAuth access tokens (WPS365 KSO-1)

USAGE:
  oauth_tools [global flags] <command> [command flags]

COMMANDS:
  token    Fetch an access token

GLOBAL FLAGS:
  -env string
        Path to .env config file (default ".env")
  -help
        Show this help message

EXAMPLES:
  oauth_tools token
  oauth_tools token -grant authorization_code
  oauth_tools token -token-only
  oauth_tools token -json
  oauth_tools -env /path/to/.env token

Run "oauth_tools <command> -help" for command-specific help.
`)
}

func main() {
	fs := flag.NewFlagSet("oauth_tools", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	envFile := fs.String("env", ".env", "Path to .env config file")
	help := fs.Bool("help", false, "Show this help message")
	fs.BoolVar(help, "h", false, "Alias for -help")

	fs.Usage = func() { printUsage(); os.Exit(2) }

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if *help {
		printUsage()
		os.Exit(0)
	}

	if fs.NArg() < 1 {
		printUsage()
		os.Exit(2)
	}

	cfg, err := config.Load(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "token":
		if err := cmd.RunToken(cfg, fs.Args()[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n", fs.Arg(0))
		printUsage()
		os.Exit(2)
	}
}
