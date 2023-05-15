package crt

import (
	"syscall"
)

func SysKill() {
	d, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return
	}
	p, err := d.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return
	}
	r, _, err := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(syscall.Getpid()))
	if r == 0 {
		return
	}
}
