package crt

import (
	"syscall"
)

func SysKill() {
	syscall.GenerateConsoleCtrlEvent(syscall.CTRL_C_EVENT, 0)
}
