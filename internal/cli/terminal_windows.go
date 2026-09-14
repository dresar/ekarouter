//go:build windows

package cli

import (
	"os"
	"syscall"
	"unsafe"
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
