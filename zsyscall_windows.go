// Code generated by 'go generate'; DO NOT EDIT.

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modUser32 = windows.NewLazySystemDLL("User32.dll")

	procGetKeyboardLayout   = modUser32.NewProc("GetKeyboardLayout")
	procLoadKeyboardLayoutW = modUser32.NewProc("LoadKeyboardLayoutW")
)

func GetKeyboardLayout(idThread uint32) (gkl uint32) {
	r0, _, _ := syscall.Syscall(procGetKeyboardLayout.Addr(), 1, uintptr(idThread), 0, 0)
	gkl = uint32(r0)
	return
}

func LoadKeyboardLayout(pwszKLID *uint16, Flags uint32) (gkl uint32) {
	r0, _, _ := syscall.Syscall(procLoadKeyboardLayoutW.Addr(), 2, uintptr(unsafe.Pointer(pwszKLID)), uintptr(Flags), 0)
	gkl = uint32(r0)
	return
}
