package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dresar/ekarouter/internal/auth"
	"github.com/dresar/ekarouter/internal/db"
)

type AppMenu struct {
	dbPath   string
	migDir   string
	port     int
	database *db.DB
	reader   *bufio.Reader
	crypto   *auth.CryptoService
}

func NewAppMenu(dbPath, migDir string, port int, secretKey string) (*AppMenu, error) {
	InitTerminal()

	if port <= 0 {
		port = 8999
	}
	if dbPath == "" {
		dbPath = "./data/ekarouter.db"
	}
	if migDir == "" {
		migDir = "migrations"
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := database.Migrate(migDir); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	var cs *auth.CryptoService
	if secretKey != "" {
		cs, _ = auth.NewCryptoService(secretKey)
	}

	return &AppMenu{
		dbPath:   dbPath,
		migDir:   migDir,
		port:     port,
		database: database,
		reader:   bufio.NewReader(os.Stdin),
		crypto:   cs,
	}, nil
}

func (m *AppMenu) Close() {
	if m.database != nil {
		_ = m.database.Close()
	}
}

func (m *AppMenu) getDashboardState() DashboardState {
	status := ProbeGateway("127.0.0.1", m.port)

	var accCount, provCount, comboCount, modelCount, proxyCount int
	if m.database != nil {
		_ = m.database.QueryRow("SELECT COUNT(*) FROM accounts WHERE enabled = 1").Scan(&accCount)
		_ = m.database.QueryRow("SELECT COUNT(*) FROM providers WHERE enabled = 1").Scan(&provCount)
		_ = m.database.QueryRow("SELECT COUNT(*) FROM routes WHERE enabled = 1").Scan(&comboCount)
		_ = m.database.QueryRow("SELECT COUNT(*) FROM models WHERE enabled = 1").Scan(&modelCount)
		_ = m.database.QueryRow("SELECT COUNT(*) FROM proxy_profiles WHERE enabled = 1").Scan(&proxyCount)
	}

	return DashboardState{
		GatewayRunning: status.Running,
		GatewayPort:    m.port,
		GatewayPID:     status.PID,
		GatewayUptime:  status.Uptime,
		DBPath:         m.dbPath,
		DBConnected:    m.database != nil,
		TotalAccounts:  accCount,
		TotalProviders: provCount,
		TotalCombos:    comboCount,
		TotalModels:    modelCount,
		TotalProxies:   proxyCount,
		TokenSaverMode: "lite",
	}
}

func (m *AppMenu) Run() {
	for {
		ClearScreen()
		RenderBanner()
		state := m.getDashboardState()
		RenderDashboard(state)

		fmt.Println(Styled("  [1] ", ColorCyan, Bold) + "🚀 Gateway Server Control " + Styled("(Start / Stop / Restart / Status)", ColorMuted))
		fmt.Println(Styled("  [2] ", ColorPurple, Bold) + "🔀 Combos & Intelligent Routing " + Styled("(MY-COMBO, cli-combo fallback chains)", ColorPurple))
		fmt.Println(Styled("  [3] ", ColorCyan, Bold) + "🌐 Providers & Accounts Pool " + Styled(fmt.Sprintf("(%d Accounts across %d Providers)", state.TotalAccounts, state.TotalProviders), ColorMuted))
		fmt.Println(Styled("  [4] ", ColorCyan, Bold) + "🛡️ Proxy Profiles & Edge Network " + Styled(fmt.Sprintf("(%d Edge Proxies configured)", state.TotalProxies), ColorMuted))
		fmt.Println(Styled("  [5] ", ColorCyan, Bold) + "🔑 Gateway API Keys Management " + Styled("(List & Generate SHA-256 Keys)", ColorMuted))
		fmt.Println(Styled("  [6] ", ColorCyan, Bold) + "⚡ Token Saver & Caveman Compression " + Styled("(RTK, Ponytail, Headroom)", ColorMuted))
		fmt.Println(Styled("  [7] ", ColorAmber, Bold) + "🛠️ AI Coding Tools Auto-Config " + Styled("(Claude Code, Cursor, Cline, Aider)", ColorAmber))
		fmt.Println(Styled("  [8] ", ColorCyan, Bold) + "📦 Database Backup & 9Router Importer")
		fmt.Println(Styled("  [9] ", ColorCyan, Bold) + "📊 Health & Diagnostic Probing")
		fmt.Println(Styled("  [0] ", ColorRed, Bold) + "🚪 Exit CLI")
		fmt.Println()

		fmt.Print(Styled("  Select option [0-9]: ", Bold, ColorWhite))
		choice := m.readLine()

		switch strings.TrimSpace(choice) {
		case "1":
			m.menuGateway()
		case "2":
			m.menuCombos()
		case "3":
			m.menuProviders()
		case "4":
			m.menuProxies()
		case "5":
			m.menuKeys()
		case "6":
			m.menuTokenSaver()
		case "7":
			m.menuTools()
		case "8":
			m.menuBackupImport()
		case "9":
			m.menuHealth()
		case "0", "q", "exit":
			fmt.Println(Styled("\n  Goodbye! Keep routing intelligently with EkaRouter.\n", ColorEmerald))
			return
		}
	}
}

func (m *AppMenu) menuGateway() {
	for {
		ClearScreen()
		RenderBanner()
		status := ProbeGateway("127.0.0.1", m.port)

		stateText := Styled("STOPPED", ColorRed, Bold)
		if status.Running {
			stateText = Styled(fmt.Sprintf("RUNNING (PID: %d, Uptime: %s, Latency: %v)", status.PID, status.Uptime, status.Latency), ColorEmerald, Bold)
		}

		fmt.Println(Styled("  ── GATEWAY SERVER CONTROL ──", Bold, ColorCyan))
		fmt.Printf("  Status:   %s\n", stateText)
		fmt.Printf("  Endpoint: %shttp://127.0.0.1:%d%s\n\n", Bold+ColorCyan, m.port, Reset)

		if status.Running {
			fmt.Println(Styled("  [1] ", ColorRed, Bold) + "🛑 Stop Gateway Daemon")
			fmt.Println(Styled("  [2] ", ColorAmber, Bold) + "🔄 Restart Gateway Daemon")
		} else {
			fmt.Println(Styled("  [1] ", ColorEmerald, Bold) + "🚀 Start Gateway Daemon (Background)")
		}
		fmt.Println(Styled("  [3] ", ColorCyan, Bold) + "🩺 Probe Live Health Endpoint")
		fmt.Println(Styled("  [0] ", ColorSlate, Bold) + "← Back to Main Menu\n")

		fmt.Print(Styled("  Select action [0-3]: ", Bold, ColorWhite))
		choice := strings.TrimSpace(m.readLine())

		switch choice {
		case "1":
			if status.Running {
				fmt.Print("  Stopping gateway... ")
				if err := StopGatewayDaemon(m.port); err != nil {
					fmt.Printf("%sFailed: %v%s\n", ColorRed, err, Reset)
				} else {
					fmt.Printf("%sStopped successfully.%s\n", ColorEmerald, Reset)
				}
			} else {
				fmt.Print("  Launching gateway daemon... ")
				newStatus, err := StartGatewayDaemon(m.dbPath, m.migDir, m.port)
				if err != nil {
					fmt.Printf("%sFailed: %v%s\n", ColorRed, err, Reset)
				} else {
					fmt.Printf("%sStarted (PID: %d)!%s\n", ColorEmerald, newStatus.PID, Reset)
				}
			}
			m.pause()
		case "2":
			if status.Running {
				fmt.Print("  Restarting gateway... ")
				_ = StopGatewayDaemon(m.port)
				time.Sleep(500 * time.Millisecond)
				newStatus, err := StartGatewayDaemon(m.dbPath, m.migDir, m.port)
				if err != nil {
					fmt.Printf("%sRestart failed: %v%s\n", ColorRed, err, Reset)
				} else {
					fmt.Printf("%sRestarted (PID: %d)!%s\n", ColorEmerald, newStatus.PID, Reset)
				}
				m.pause()
			}
		case "3":
			s := ProbeGateway("127.0.0.1", m.port)
			if s.Running {
				fmt.Printf("\n  %s✓ Health OK: Uptime %s, Latency %v%s\n", ColorEmerald, s.Uptime, s.Latency, Reset)
			} else {
				fmt.Printf("\n  %s✗ Gateway is not responding on port %d%s\n", ColorRed, m.port, Reset)
			}
			m.pause()
		case "0":
			return
		}
	}
}

func (m *AppMenu) menuCombos() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── COMBOS & INTELLIGENT ROUTING CHAINS ──", Bold, ColorPurple))

	routes, err := m.database.Query("SELECT id, name, strategy FROM routes WHERE enabled = 1")
	if err != nil {
		fmt.Printf("  %sError: %v%s\n", ColorRed, err, Reset)
		m.pause()
		return
	}
	defer routes.Close()

	for routes.Next() {
		var rID, rName, rStrategy string
		if err := routes.Scan(&rID, &rName, &rStrategy); err == nil {
			fmt.Printf("\n  %s🔀 Combo: %s%s %s(Strategy: %s)%s\n", Bold+ColorCyan, rName, Reset, ColorMuted, rStrategy, Reset)

			items, err := m.database.Query(`
SELECT ri.priority, ri.provider_id, COALESCE(m.external_name, ri.model_id)
FROM route_items ri
LEFT JOIN models m ON m.id = ri.model_id
WHERE ri.route_id = ? AND ri.enabled = 1
ORDER BY ri.priority ASC
LIMIT 16`, rID)

			if err == nil {
				for items.Next() {
					var prio int
					var prov, model string
					if err := items.Scan(&prio, &prov, &model); err == nil {
						fmt.Printf("     #%02d → %s%-16s%s : %s\n", prio, ColorSlate, prov, Reset, model)
					}
				}
				items.Close()
			}
		}
	}
	fmt.Println()
	m.pause()
}

func (m *AppMenu) menuProviders() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── REGISTERED PROVIDERS & ACCOUNTS (182 TOTAL) ──", Bold, ColorCyan))

	rows, err := m.database.Query(`
SELECT p.id, p.name, p.kind, COUNT(a.id)
FROM providers p
LEFT JOIN accounts a ON a.provider_id = p.id AND a.enabled = 1
WHERE p.enabled = 1
GROUP BY p.id
ORDER BY COUNT(a.id) DESC
LIMIT 24`)

	if err != nil {
		fmt.Printf("  %sQuery error: %v%s\n", ColorRed, err, Reset)
		m.pause()
		return
	}
	defer rows.Close()

	fmt.Printf("  %-24s %-12s %-10s\n", Styled("PROVIDER", Bold), Styled("KIND", Bold), Styled("ACCOUNTS", Bold))
	fmt.Println("  " + strings.Repeat("─", 52))

	for rows.Next() {
		var id, name, kind string
		var count int
		if err := rows.Scan(&id, &name, &kind, &count); err == nil {
			fmt.Printf("  %-24s %-12s %s%d accounts%s\n", id, kind, Bold+ColorEmerald, count, Reset)
		}
	}
	fmt.Println()
	m.pause()
}

func (m *AppMenu) menuProxies() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── PROXY POOLS & EDGE NETWORKING ──", Bold, ColorCyan))

	rows, err := m.database.Query("SELECT id, name, scheme, host, port, enabled FROM proxy_profiles ORDER BY id LIMIT 20")
	if err != nil {
		fmt.Printf("  %sError: %v%s\n", ColorRed, err, Reset)
		m.pause()
		return
	}
	defer rows.Close()

	fmt.Printf("  %-20s %-8s %-32s %-8s\n", Styled("NAME", Bold), Styled("SCHEME", Bold), Styled("HOST", Bold), Styled("STATUS", Bold))
	fmt.Println("  " + strings.Repeat("─", 72))

	for rows.Next() {
		var id, name, scheme, host string
		var port, enabled int
		if err := rows.Scan(&id, &name, &scheme, &host, &port, &enabled); err == nil {
			st := Styled("Active", ColorEmerald)
			if enabled == 0 {
				st = Styled("Disabled", ColorRed)
			}
			fmt.Printf("  %-20s %-8s %-32s %s\n", name, scheme, fmt.Sprintf("%s:%d", host, port), st)
		}
	}
	fmt.Println()
	m.pause()
}

func (m *AppMenu) menuKeys() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── GATEWAY API KEYS MANAGEMENT ──", Bold, ColorCyan))

	rows, err := m.database.Query("SELECT id, name, prefix, enabled, created_at FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		fmt.Printf("  %sError: %v%s\n", ColorRed, err, Reset)
		m.pause()
		return
	}
	defer rows.Close()

	fmt.Printf("  %-16s %-32s %-10s\n", Styled("NAME", Bold), Styled("KEY PREFIX", Bold), Styled("STATUS", Bold))
	fmt.Println("  " + strings.Repeat("─", 60))

	for rows.Next() {
		var id, name, prefix string
		var enabled int
		var createdAt string
		if err := rows.Scan(&id, &name, &prefix, &enabled, &createdAt); err == nil {
			st := Styled("Active", ColorEmerald)
			if enabled == 0 {
				st = Styled("Disabled", ColorRed)
			}
			fmt.Printf("  %-16s %-32s %s\n", name, prefix+"...", st)
		}
	}
	fmt.Println()
	m.pause()
}

func (m *AppMenu) menuTokenSaver() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── TOKEN SAVER COMPRESSION & CAVEMAN ──", Bold, ColorAmber))
	fmt.Println("  Status: " + Badge("ACTIVE", ColorWhite, BgGreen) + " • Mode: " + Bold + ColorAmber + "LITE" + Reset)
	fmt.Println("\n  Active Compression Filters:")
	fmt.Println("    • RTK (Redundant Tool Output Stripper - 12 regex filters)")
	fmt.Println("    • Caveman Protocol (Compresses natural language without semantic loss)")
	fmt.Println("    • Ponytail & Headroom Safety Token Window Limits")
	fmt.Println("    • Pxpipe Context Optimization")
	fmt.Println()
	m.pause()
}

func (m *AppMenu) menuTools() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── 🛠️ AI CODING TOOLS AUTO-CONFIG ──", Bold, ColorAmber))

	apiKey := GetDefaultGatewayKey(m.database.DB)

	claude := BuildClaudeCodeConfig(m.port, apiKey)
	cursor := BuildCursorConfig(m.port, apiKey)
	cline := BuildClineConfig(m.port, apiKey)
	aider := BuildAiderConfig(m.port, apiKey)

	fmt.Println(Styled("  [1] Claude Code CLI Config", Bold, ColorCyan))
	fmt.Println(Styled("  [2] Cursor IDE Integration", Bold, ColorCyan))
	fmt.Println(Styled("  [3] Cline / Roo-Code Setup", Bold, ColorCyan))
	fmt.Println(Styled("  [4] Aider CLI Command", Bold, ColorCyan))
	fmt.Println(Styled("  [0] ← Back", ColorSlate))
	fmt.Println()

	fmt.Print(Styled("  Select tool [0-4]: ", Bold, ColorWhite))
	choice := strings.TrimSpace(m.readLine())

	switch choice {
	case "1":
		ClearScreen()
		fmt.Println(Styled("  ── CLAUDE CODE CLI SETUP ──\n", Bold, ColorCyan))
		for title, snip := range claude.Snippets {
			fmt.Printf("  %s%s:%s\n%s\n\n", Bold, title, Reset, snip)
		}
		m.pause()
	case "2":
		ClearScreen()
		fmt.Println(Styled("  ── CURSOR IDE SETUP ──\n", Bold, ColorCyan))
		for title, snip := range cursor.Snippets {
			fmt.Printf("  %s%s:%s\n%s\n\n", Bold, title, Reset, snip)
		}
		m.pause()
	case "3":
		ClearScreen()
		fmt.Println(Styled("  ── CLINE SETUP ──\n", Bold, ColorCyan))
		for title, snip := range cline.Snippets {
			fmt.Printf("  %s%s:%s\n%s\n\n", Bold, title, Reset, snip)
		}
		m.pause()
	case "4":
		ClearScreen()
		fmt.Println(Styled("  ── AIDER CLI COMMAND ──\n", Bold, ColorCyan))
		for title, snip := range aider.Snippets {
			fmt.Printf("  %s%s:%s\n%s\n\n", Bold, title, Reset, snip)
		}
		m.pause()
	}
}

func (m *AppMenu) menuBackupImport() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── DATABASE BACKUP & 9ROUTER IMPORTER ──", Bold, ColorCyan))
	fmt.Println(Styled("  [1] ", ColorCyan, Bold) + "Import 9Router JSON Backup")
	fmt.Println(Styled("  [2] ", ColorEmerald, Bold) + "Export SQLite Snapshot")
	fmt.Println(Styled("  [0] ", ColorSlate, Bold) + "← Back\n")

	fmt.Print(Styled("  Select action [0-2]: ", Bold, ColorWhite))
	choice := strings.TrimSpace(m.readLine())

	switch choice {
	case "1":
		fmt.Print("\n  Enter path to backup JSON file: ")
		p := strings.TrimSpace(m.readLine())
		if p == "" {
			return
		}
		fmt.Print("  Importing... ")
		stats, err := db.Import9RouterBackup(m.database, m.crypto, p)
		if err != nil {
			fmt.Printf("%sImport failed: %v%s\n", ColorRed, err, Reset)
		} else {
			fmt.Printf("%sImported %d accounts, %d models, %d combos!%s\n", ColorEmerald, stats.Accounts, stats.Models, stats.Routes, Reset)
		}
		m.pause()
	case "2":
		backupPath := fmt.Sprintf("./data/backup-%d.db", time.Now().Unix())
		fmt.Printf("  Backing up SQLite database to %s... ", backupPath)
		if err := m.database.Backup(backupPath); err != nil {
			fmt.Printf("%sBackup failed: %v%s\n", ColorRed, err, Reset)
		} else {
			fmt.Printf("%sSuccess!%s\n", ColorEmerald, Reset)
		}
		m.pause()
	}
}

func (m *AppMenu) menuHealth() {
	ClearScreen()
	RenderBanner()
	fmt.Println(Styled("  ── HEALTH & DIAGNOSTICS ──", Bold, ColorCyan))

	gwStatus := ProbeGateway("127.0.0.1", m.port)
	if gwStatus.Running {
		fmt.Printf("  %s[✓] Gateway HTTP API:%s Online on :%d (Uptime: %s, Latency: %v)\n", ColorEmerald, Reset, m.port, gwStatus.Uptime, gwStatus.Latency)
	} else {
		fmt.Printf("  %s[✗] Gateway HTTP API:%s Offline on :%d\n", ColorRed, Reset, m.port)
	}

	if m.database != nil {
		var n int
		err := m.database.QueryRow("SELECT 1").Scan(&n)
		if err == nil {
			fmt.Printf("  %s[✓] SQLite Database:%s  Operational (WAL Mode, zero-CGO)\n", ColorEmerald, Reset)
		} else {
			fmt.Printf("  %s[✗] SQLite Database:%s  Error: %v\n", ColorRed, Reset, err)
		}
	}

	fmt.Println()
	m.pause()
}

func (m *AppMenu) readLine() string {
	text, _ := m.reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func (m *AppMenu) pause() {
	fmt.Print(Styled("\n  Press Enter to continue...", ColorMuted))
	_, _ = m.reader.ReadString('\n')
}
