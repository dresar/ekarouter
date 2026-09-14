package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dresar/ekarouter/internal/db"
)

func HandleSubcommands(args []string, dbPath, migDir string, port int) bool {
	if len(args) == 0 {
		return false
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "cli", "menu", "interactive", "ui":
		menu, err := NewAppMenu(dbPath, migDir, port, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start interactive CLI: %v\n", err)
			os.Exit(1)
		}
		defer menu.Close()
		menu.Run()
		return true

	case "status":
		status := ProbeGateway("127.0.0.1", port)
		if status.Running {
			fmt.Printf("%s[ONLINE]%s Gateway running on http://127.0.0.1:%d (PID: %d, Uptime: %s, Latency: %v)\n",
				ColorEmerald, Reset, port, status.PID, status.Uptime, status.Latency)
		} else {
			fmt.Printf("%s[OFFLINE]%s Gateway is not running on port %d\n", ColorRed, Reset, port)
		}
		return true

	case "start":
		fmt.Printf("Starting EkaRouter gateway daemon on port %d...\n", port)
		status, err := StartGatewayDaemon(dbPath, migDir, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%sSuccess!%s Gateway running at http://127.0.0.1:%d (PID: %d)\n", ColorEmerald, Reset, status.Port, status.PID)
		return true

	case "stop":
		fmt.Printf("Stopping EkaRouter gateway daemon on port %d...\n", port)
		if err := StopGatewayDaemon(port); err != nil {
			fmt.Fprintf(os.Stderr, "failed to stop: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%sGateway stopped.%s\n", ColorEmerald, Reset)
		return true

	case "restart":
		_ = StopGatewayDaemon(port)
		status, err := StartGatewayDaemon(dbPath, migDir, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to restart: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%sGateway restarted!%s (PID: %d)\n", ColorEmerald, Reset, status.PID)
		return true

	case "tools", "tool":
		handleToolsCommand(args[1:], dbPath, port)
		return true

	case "combos", "routes":
		handleCombosCommand(dbPath)
		return true

	case "providers", "accounts":
		handleProvidersCommand(dbPath)
		return true
	}

	return false
}

func handleCombosCommand(dbPath string) {
	database, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	routes, err := database.Query("SELECT id, name, strategy FROM routes WHERE enabled = 1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "query routes: %v\n", err)
		os.Exit(1)
	}
	defer routes.Close()

	for routes.Next() {
		var rID, rName, rStrategy string
		if err := routes.Scan(&rID, &rName, &rStrategy); err == nil {
			fmt.Printf("Combo [%s] (Strategy: %s)\n", rName, rStrategy)
			items, err := database.Query(`
SELECT ri.priority, ri.provider_id, COALESCE(m.external_name, ri.model_id)
FROM route_items ri
LEFT JOIN models m ON m.id = ri.model_id
WHERE ri.route_id = ? AND ri.enabled = 1
ORDER BY ri.priority ASC
LIMIT 10`, rID)

			if err == nil {
				for items.Next() {
					var prio int
					var prov, model string
					if err := items.Scan(&prio, &prov, &model); err == nil {
						fmt.Printf("  #%02d -> %-16s : %s\n", prio, prov, model)
					}
				}
				items.Close()
			}
			fmt.Println()
		}
	}
}

func handleProvidersCommand(dbPath string) {
	database, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	rows, err := database.Query(`
SELECT p.id, p.kind, COUNT(a.id)
FROM providers p
LEFT JOIN accounts a ON a.provider_id = p.id AND a.enabled = 1
WHERE p.enabled = 1
GROUP BY p.id
ORDER BY COUNT(a.id) DESC
LIMIT 15`)

	if err != nil {
		fmt.Fprintf(os.Stderr, "query providers: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Printf("%-24s %-12s %-10s\n", "PROVIDER", "KIND", "ACCOUNTS")
	fmt.Println(strings.Repeat("-", 48))
	for rows.Next() {
		var id, kind string
		var count int
		if err := rows.Scan(&id, &kind, &count); err == nil {
			fmt.Printf("%-24s %-12s %d accounts\n", id, kind, count)
		}
	}
}

func handleToolsCommand(subArgs []string, dbPath string, port int) {
	database, err := db.Open(dbPath)
	apiKey := "sk-7016dd129191d903-u8emvi-c0b721b9"
	if err == nil {
		apiKey = GetDefaultGatewayKey(database.DB)
		_ = database.Close()
	}

	target := "all"
	if len(subArgs) > 0 {
		target = strings.ToLower(subArgs[0])
	}

	switch target {
	case "claude":
		cfg := BuildClaudeCodeConfig(port, apiKey)
		fmt.Println(cfg.Snippets["PowerShell"])
	case "cursor":
		cfg := BuildCursorConfig(port, apiKey)
		fmt.Println(cfg.Snippets["settings.json"])
	case "cline":
		cfg := BuildClineConfig(port, apiKey)
		fmt.Println(cfg.Snippets["cline_custom_modes.json"])
	case "aider":
		cfg := BuildAiderConfig(port, apiKey)
		fmt.Println(cfg.Snippets["Terminal Command"])
	default:
		fmt.Println("Available tools: claude, cursor, cline, aider")
	}
}
