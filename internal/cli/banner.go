package cli

import (
	"fmt"
	"strings"
	"time"
)

type DashboardState struct {
	GatewayRunning bool
	GatewayPort    int
	GatewayPID     int
	GatewayUptime  string
	DBPath         string
	DBConnected    bool
	TotalAccounts  int
	TotalProviders int
	TotalCombos    int
	TotalModels    int
	TotalProxies   int
	TokenSaverMode string
}

func RenderBanner() {
	lines := []string{
		`  ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗   ██╗████████╗███████╗██████╗ `,
		`  ██╔════╝██║ ██╔╝██╔══██╗██╔══██╗██╔═══██╗██║   ██║╚══██╔══╝██╔════╝██╔══██╗`,
		`  █████╗  █████═╝ ███████║██████╔╝██║   ██║██║   ██║   ██║   █████╗  ██████╔╝`,
		`  ██╔══╝  ██╔═██╗ ██╔══██║██╔══██╗██║   ██║██║   ██║   ██║   ██╔══╝  ██╔══██╗`,
		`  ███████╗██║ ╚██╗██║  ██║██║  ██║╚██████╔╝╚██████╔╝   ██║   ███████╗██║  ██║`,
		`  ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚═╝  ╚═╝`,
	}

	palette := []string{ColorCyan, ColorEmerald, ColorBlue, ColorPurple, ColorAmber, ColorCyan}

	fmt.Println()
	for i, l := range lines {
		c := palette[i%len(palette)]
		fmt.Printf("%s%s%s%s\n", Bold, c, l, Reset)
	}

	badge := fmt.Sprintf("  %s %s %s %s",
		Badge("v1.0.0", ColorWhite, BgCyan),
		Badge("ZERO-CGO", ColorWhite, BgGreen),
		Badge("AES-256", ColorWhite, BgPurple),
		Styled("Lightweight High-Performance Go AI Gateway", ColorSlate, Italic),
	)
	fmt.Println(badge)
	fmt.Println()
}

func RenderDashboard(s DashboardState) {
	gwBadge := Badge("● ONLINE", ColorWhite, BgGreen)
	if !s.GatewayRunning {
		gwBadge = Badge("○ STOPPED", ColorWhite, BgDark)
	}

	var statusLines []string

	gwInfo := fmt.Sprintf("  %s  Endpoint: %shttp://127.0.0.1:%d%s",
		gwBadge, Bold+ColorCyan, s.GatewayPort, Reset)
	if s.GatewayRunning && s.GatewayPID > 0 {
		gwInfo += fmt.Sprintf("  %s(PID: %d, Uptime: %s)%s", ColorMuted, s.GatewayPID, s.GatewayUptime, Reset)
	}
	statusLines = append(statusLines, gwInfo)

	dbStatus := Styled("Connected (WAL Mode)", ColorEmerald)
	if !s.DBConnected {
		dbStatus = Styled("Disconnected", ColorRed)
	}
	statusLines = append(statusLines, fmt.Sprintf("  %sDatabase:%s %s %s•%s %s%d Accounts%s across %s%d Providers%s",
		ColorSlate, Reset, dbStatus, ColorMuted, Reset, Bold+ColorWhite, s.TotalAccounts, Reset, Bold+ColorWhite, s.TotalProviders, Reset))

	statusLines = append(statusLines, fmt.Sprintf("  %sRouting:%s  %s%d Active Combos%s %s(MY-COMBO, cli-combo)%s • %s%d Models%s",
		ColorSlate, Reset, Bold+ColorPurple, s.TotalCombos, Reset, ColorMuted, Reset, Bold+ColorWhite, s.TotalModels, Reset))

	statusLines = append(statusLines, fmt.Sprintf("  %sNetwork:%s  %s%d Edge Proxies%s • SOCKS5/HTTP Egress Protection",
		ColorSlate, Reset, Bold+ColorCyan, s.TotalProxies, Reset))

	tsMode := s.TokenSaverMode
	if tsMode == "" {
		tsMode = "lite"
	}
	statusLines = append(statusLines, fmt.Sprintf("  %sSaver:%s    %s⚡ Token Saver Active%s • Mode: %s%s%s • RTK + Caveman Filters",
		ColorSlate, Reset, ColorAmber, Reset, Bold+ColorAmber, strings.ToUpper(tsMode), Reset))

	fmt.Print(Box("SYSTEM OVERVIEW", statusLines, 84, ColorSlate))
	fmt.Println()
}

func FormatUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
