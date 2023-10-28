//go:build windows
// +build windows

package main

const ()

//https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getkeyboardlayout
//sys GetKeyboardLayout(idThread uint32) (gkl uint32) = User32.GetKeyboardLayout

//https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-loadkeyboardlayouta
// sys LoadKeyboardLayout(pwszKLID []byte, Flags uint32) (gkl uint32) = User32.LoadKeyboardLayoutA
//sys LoadKeyboardLayout(pwszKLID *uint16, Flags uint32)(gkl uint32) = User32.LoadKeyboardLayoutW
