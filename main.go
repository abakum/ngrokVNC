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

	TO         = time.Second * 10
	portRFB    = "5900"
	portViewer = 5500
	VNC        = map[string]string{"name": ""}
	TightVNC   = map[string]string{
		"name":     "TightVNC",
		"server":   "tvnserver.exe",
		"viewer":   "tvnviewer.exe",
		"path":     "",
		"services": "tvnserver",
		"kill":     "-shutdown",
	}
	UltraVNC = map[string]string{
		"name":     "UltraVNC",
		"server":   "winvnc.exe",
		"viewer":   "vncviewer.exe",
		"path":     "",
		"services": "uvnc_service",
		"kill":     "-kill",
	}
	VNCs = []map[string]string{TightVNC, UltraVNC}
	serverExe,
	viewerExe,
	control string
	AcceptRfbConnections = true
	proxy, localListen   bool
	k                    registry.Key
)

func main() {
	cwd, err := os.Getwd()
	if err == nil {
		for _, xVNC := range VNCs {
			xVNC["path"] = filepath.Join(cwd, "..", xVNC["name"])
			if stat, err := os.Stat(xVNC["path"]); err == nil && stat.IsDir() {
				VNC = xVNC
				break
			}
		}
	}

	key := `SOFTWARE\Classes\VncViewer.Config\DefaultIcon`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE|registry.SET_VALUE)
	key = ""
	if err == nil {
		old, _, err := k.GetStringValue(key)
		if err == nil {
			for _, xVNC := range VNCs {
				if !strings.Contains(old, xVNC["name"]) {
					continue
				}
				dir := filepath.Dir(old)
				if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
					VNC = xVNC
					VNC["path"] = dir
					break
				}
			}
		} else {
			PrintOk(key, err)
		}
		k.Close()
	} else {
		PrintOk(key, err)
	}
	if VNC["name"] == "" {
		letf.Println("not found vnc")
		return
	}
	li.Println(VNC["name"], VNC["path"])
	serverExe = filepath.Join(VNC["path"], VNC["server"])
	viewerExe = filepath.Join(VNC["path"], VNC["viewer"])

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
				// :host[::port] [password] as ngrok~proxy~LAN viewer connect mode
				// :host[:screen] [password] as ngrok~proxy~LAN viewer connect mode
				// : [password] as ngrok viewer connect mode
				// host as host:0 as host: as host::5900 as host::
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
func usage() {
	li.Println("Run - запусти")
	li.Println("`ngrokVNC [::port]`")
	li.Println("When there is no ngrok tunnel it will be created  - когда ngrok туннеля нет он создатся")
	li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC : [password]`")
	li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC host[::port] [password]`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC 0`")
	li.Println("This will create a ngrok tunnel - это создаст туннель")
	li.Println("The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через туннель")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC -0`")
	li.Println("The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN")
	li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC -host`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC -`")
	li.Println("the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти")
	li.Println("\t`ngrokVNC 0`")
}
