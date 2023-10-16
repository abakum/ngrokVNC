//go:build windows
// +build windows

package main

const (
	HWND_BROADCAST            uint32 = 0x0000FFFF
	WM_INPUTLANGCHANGEREQUEST uint32 = 0x00000050
)

//https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getkeyboardlayout
//sys GetKeyboardLayout(idThread uint32) (gkl uint32) = User32.GetKeyboardLayout

//https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-loadkeyboardlayouta
//sys LoadKeyboardLayout(pwszKLID []byte, Flags uint32) (gkl uint32) = User32.LoadKeyboardLayoutA

//https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-postmessagea
//sys PostMessage (hWnd uint32, Msg uint32, wParam uint32, lParam uint32) (err error)= User32.PostMessageA
