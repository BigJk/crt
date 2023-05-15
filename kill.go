//go:build !windows && !js
// +build !windows,!js

package crt

import "syscall"

func SysKill() {
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
