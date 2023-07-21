package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string

	TO                   = time.Second * 10
	portRFB              = "5900"
	portViewer           = 5500
	TightVNC             = "TightVNC"
	tvnserver            = "tvnserver.exe"
	tvnviewer            = "tvnviewer.exe"
	AcceptRfbConnections = true
)

func main() {
	cwd, err := os.Getwd()
	if err == nil {
		TightVNC = filepath.Join(cwd, "..", TightVNC)
	}
	key := `SOFTWARE\Classes\VncViewer.Config\DefaultIcon`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE|registry.SET_VALUE)
	PrintOk(key, err)
	if err == nil {
		key = ""
		old, _, err := k.GetStringValue(key)
		PrintOk(key, err)
		if err == nil && strings.Contains(old, `\`) {
			TightVNC = filepath.Dir(old)
		}
		k.Close()
	}
	li.Println("TightVNC", TightVNC)
	tvnserver = filepath.Join(TightVNC, tvnserver)
	tvnviewer = filepath.Join(TightVNC, tvnviewer)

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then local mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate local mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY)
	// NGROK_API_KEY = ""

	if len(os.Args) > 1 {
		_, err := strconv.Atoi(os.Args[1])
		if err != nil {
			switch {
			case os.Args[1] == "-":
				// - try connect server to viewer via ngrok (revers)
				serverNgrok()
			case strings.HasPrefix(os.Args[1], "-"):
				// - try connect server to viewer via LAN (revers)
				serverLAN()
			case strings.HasPrefix(os.Args[1], "::"):
				// :: as ::5900
				server()
			default:
				// host[::port] [password] as LAN viewer connect mode
				// host[:screen] [password] as LAN viewer connect mode
				// : [password] as ngrok viewer connect mode
				// host as host:0 as host::5900
				viewer()
			}
			return
		}
		// -port as LAN viewer listen mode
		// port as ngrok viewer listen mode
		// 0  as 5500
		viewerl()
		return
	}
	// as GUI or reg RfbPort
	server()
}

func abs(s string) string {
	if strings.HasPrefix(s, "-") {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		return strings.TrimPrefix(s, "-")
	}
	return s
}
