package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string

	port      = "5900"
	TightVNC  = "TightVNC"
	tvnserver = "tvnserver.exe"
	tvnviewer = "tvnviewer.exe"
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

	li.Println("Without params and ngrok not exist - без параметров и ngrok туннель отсутствует:")
	li.Println("VNC server mode - экран VNC ожидает подключения VNC наблюдателя")
	li.Println("To view over the tunnel on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("`ngrokVNC :`")
	li.Println("To view over the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("`ngrokVNC host`")
	li.Println()
	li.Println("Without params and ngrok exist - без параметров и ngrok туннель существует:")
	li.Println("VNC server connect to viewer mode - экран VNC подключается к ожидающему наблюдателю")
	li.Println("On the other side was launched - на другой стороне был запущен")
	li.Println("`ngrokVNC 0`")
	li.Println()
	li.Println("To connect the VNC server to the viewer over the ngrok, run - для подключения экрана VNC к наблюдателю через туннель запусти")
	li.Println("`ngrokVNC ::`")
	li.Println("On the other side run - на другой стороне запусти")
	li.Println("`ngrokVNC 0`")

	if len(os.Args) > 1 {
		_, err := strconv.Atoi(os.Args[1])
		if err != nil {
			// :: try connect server to viewer over ngrok
			if os.Args[1] == "::" {
				serverc()
				return
			}
			// host[::port] [password] as LAN viewer connect mode
			// host[:screen] [password] as LAN viewer connect mode
			// : [password] as ngrok viewer connect mode
			// host as host:0 as host::5900
			viewer()
			return
		}
		// -port as LAN viewer listen mode
		// port as ngrok viewer listen mode
		// 0  as 5500
		viewerl()
		return
	}
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
