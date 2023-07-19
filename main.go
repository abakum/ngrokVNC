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

	TO        = time.Second * 10
	port      = "5900"
	p         = 5500
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

	li.Println("Run - запусти")
	li.Println("`ngrokVNC`")
	li.Println("When there is no ngrok tunnel it will be created  - когда ngrok туннеля нет он создатся")
	li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC :`")
	li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC host`")
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

	if len(os.Args) > 1 {
		_, err := strconv.Atoi(os.Args[1])
		if err != nil {
			// - try connect server to viewer over ngrok (revers)
			if os.Args[1] == "-" {
				serverNgrok()
				return
			}
			// -host try connect server to viewer over LAN (revers)
			if strings.HasPrefix(os.Args[1], "-") {
				serverLAN()
				return
			}
			// host[::port] [password] as LAN viewer connect mode
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
