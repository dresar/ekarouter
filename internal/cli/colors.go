package cli

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	ColorCyan    = "\033[38;2;6;182;212m"
	ColorEmerald = "\033[38;2;16;185;129m"
	ColorPurple  = "\033[38;2;168;85;247m"
	ColorAmber   = "\033[38;2;245;158;11m"
	ColorRed     = "\033[38;2;239;68;68m"
	ColorBlue    = "\033[38;2;59;130;246m"
	ColorSlate   = "\033[38;2;148;163;184m"
	ColorMuted   = "\033[38;2;100;116;139m"
	ColorWhite   = "\033[38;2;248;250;252m"

	BgDark   = "\033[48;2;15;23;42m"
	BgPurple = "\033[48;2;88;28;135m"
	BgCyan   = "\033[48;2;14;116;144m"
	BgGreen  = "\033[48;2;6;95;70m"
)

func InitTerminal() {
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	if r1, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); r1 != 0 {
		const enableVirtualTerminalProcessing = 0x0004
		_, _, _ = setConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
	}
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func Styled(text string, colors ...string) string {
	var sb strings.Builder
	for _, c := range colors {
		sb.WriteString(c)
	}
	sb.WriteString(text)
	sb.WriteString(Reset)
	return sb.String()
}

func Badge(text, fg, bg string) string {
	return fmt.Sprintf("%s%s %s %s", bg, fg, text, Reset)
}

func Box(title string, lines []string, width int, borderColor string) string {
	if width < 40 {
		width = 76
	}

	var sb strings.Builder
	titleLen := len(title)
	remain := width - titleLen - 4
	if remain < 2 {
		remain = 2
	}

	sb.WriteString(borderColor)
	sb.WriteString("╭─ ")
	sb.WriteString(Reset)
	sb.WriteString(Bold)
	sb.WriteString(ColorWhite)
	sb.WriteString(title)
	sb.WriteString(Reset)
	sb.WriteString(borderColor)
	sb.WriteString(" " + strings.Repeat("─", remain) + "╮\n")

	for _, l := range lines {
		sb.WriteString("│ ")
		sb.WriteString(Reset)
		sb.WriteString(l)
		sb.WriteString(borderColor)

		visibleLen := stripAnsiLength(l)
		padding := width - visibleLen - 1
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString("│\n")
	}

	sb.WriteString("╰" + strings.Repeat("─", width+1) + "╯\n")
	sb.WriteString(Reset)
	return sb.String()
}

func stripAnsiLength(s string) int {
	inEscape := false
	length := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}
