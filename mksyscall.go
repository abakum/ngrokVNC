//go:build windows
// +build windows

package main

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go syscall.go
