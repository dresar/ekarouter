package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dresar/ekarouter/internal/app"
	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/cli"
	"github.com/dresar/ekarouter/internal/config"
	"github.com/dresar/ekarouter/internal/db"
)

var Version = "1.0.0"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		if os.Args[1] != "serve" && os.Args[1] != "server" {
			if cli.HandleSubcommands(os.Args[1:], cfg.DatabasePath, "migrations", cfg.Port) {
				return
			}
		}
	} else if len(os.Args) == 1 {
		if cli.HandleSubcommands([]string{"cli"}, cfg.DatabasePath, "migrations", cfg.Port) {
			return
		}
	}

	hostFlag := flag.String("host", "", "HTTP bind address")
	portFlag := flag.Int("port", 0, "HTTP port")
	dbFlag := flag.String("db", "", "SQLite database path")
	migFlag := flag.String("migrations", "migrations", "Path to SQL migrations folder")
	importFlag := flag.String("import", "", "Path to 9router backup JSON file to import")
	backupFlag := flag.String("backup", "", "Run database backup to file and exit")
	versionFlag := flag.Bool("version", false, "Print application version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("EkaRouter v%s\n", Version)
		return
	}

	if *hostFlag != "" {
		cfg.Host = *hostFlag
	}
	if *portFlag > 0 {
		cfg.Port = *portFlag
	}
	if *dbFlag != "" {
		cfg.DatabasePath = *dbFlag
	}

	if *backupFlag != "" {
		database, err := db.Open(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open database for backup: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		if err := database.Backup(*backupFlag); err != nil {
			fmt.Fprintf(os.Stderr, "database backup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("database successfully backed up to: %s\n", *backupFlag)
		return
	}

	if *importFlag != "" {
		database, err := db.Open(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open database for import: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		if err := database.Migrate(*migFlag); err != nil {
			fmt.Fprintf(os.Stderr, "database migration failed: %v\n", err)
			os.Exit(1)
		}

		crypto, err := auth.NewCryptoService(cfg.SecretKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "crypto error: %v\n", err)
			os.Exit(1)
		}

		stats, err := db.Import9RouterBackup(database, crypto, *importFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully imported from %s:\n", *importFlag)
		fmt.Printf("  Providers:    %d\n", stats.Providers)
		fmt.Printf("  Accounts:     %d\n", stats.Accounts)
		fmt.Printf("  Credentials:  %d\n", stats.Credentials)
		fmt.Printf("  Models:       %d\n", stats.Models)
		fmt.Printf("  Routes:       %d\n", stats.Routes)
		fmt.Printf("  Route Items:  %d\n", stats.RouteItems)
		fmt.Printf("  Proxy Pools:  %d\n", stats.ProxyPools)
		fmt.Printf("  API Keys:     %d\n", stats.APIKeys)
		fmt.Printf("  Settings:     %d\n", stats.Settings)
		return
	}

	application, err := app.Setup(cfg, *migFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "application setup error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("EkaRouter v%s listening on %s:%d\n", Version, cfg.Host, cfg.Port)

	if err := application.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("EkaRouter stopped gracefully.")
}
