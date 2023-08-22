package main

import (
	_ "embed"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
	"gopkg.in/ini.v1"
)

const (
	CportRFB    = "5900"
	CportViewer = 5500
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string
	//go:embed uvnc.pkey.txt
	pkey []byte

	keyFN       = "20230722_Viewer_ClientAuth.pkey"
	errC        error
	TO          = time.Second * 60
	portRFB     = CportRFB
	portViewer  = CportViewer
	RportRFB    = ""
	RportViewer = 0
	PportRFB    = CportRFB
	PportViewer = CportViewer
	VNC         = map[string]string{"name": ""}
	TightVNC    = map[string]string{
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
	TurboVNC = map[string]string{
		"name":     "TurboVNC",
		"server":   "",
		"viewer":   "vncviewerw.bat",
		"path":     "",
		"services": "",
		"kill":     "-kill",
	}
	RealVNC = map[string]string{
		"name":     "RealVNC",
		"server":   "",
		"viewer":   "vncviewer.exe",
		"path":     "",
		"services": "",
		"kill":     "",
	}
	VNCs = []map[string]string{TightVNC, UltraVNC, TurboVNC, RealVNC}
	serverExe,
	viewerExe,
	control,
	connect,
	publicURL,
	forwardsTo,
	listen,
	inLAN,
	ip,
	ifs,
	ultravnc,
	DSMPlugin string
	UseDSMPlugin         = "0"
	id                   = "0"
	AcceptRfbConnections = true
	proxy,
	proxy2,
	rProxy,
	rProxy2,
	localListen,
	plus,
	plus2,
	reload,
	first bool
	k       registry.Key
	tcp     *url.URL
	ips     []string
	servers int
)

func main() {
	if len(os.Args) == 1 {
		usage()
	}
	executable, _ := os.Executable()
	imagename := filepath.Base(executable)
	first = strings.Count(taskList("imagename eq "+imagename), imagename) == 1

	ips = interfaces()
	if len(ips) == 0 {
		letf.Println("not connected")
		return
	}
	ifs = strings.Join(ips, ",")

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
	if VNC["path"] == "" {
		letf.Println("not found VNC viewer")
		return
	}
	li.Println(VNC["name"], VNC["path"])
	serverExe = filepath.Join(VNC["path"], VNC["server"])
	viewerExe = filepath.Join(VNC["path"], VNC["viewer"])
	servers = strings.Count(taskList("imagename eq "+VNC["server"]), VNC["server"])
	// letf.Println("servers", servers)
	if VNC["name"] == "UltraVNC" {
		ultravnc = filepath.Join(VNC["path"], "ultravnc.ini")
		iniFile, err := ini.Load(ultravnc)
		if err == nil {
			section := iniFile.Section("admin")
			DSMPlugin = section.Key("DSMPlugin").String()
			if DSMPlugin != "" { //section.Key("UseDSMPlugin").String() == "1" &&
				UseDSMPlugin = "1"
			}
		} else {
			letf.Println("error read", ultravnc)
		}
		keyFN = filepath.Join(VNC["path"], keyFN)
		_, err = os.Stat(keyFN)
		if err != nil {
			PrintOk(keyFN, os.WriteFile(keyFN, pkey, 0666))
		}
	}

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //if emty then LAN mode
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                   // emulate LAN mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY)
	// NGROK_API_KEY = "" // emulate no crypt

	args := os.Args[:]
	for k, v := range args {
		args[k] = strings.ReplaceAll(v, ";", ":")
	}

	p5ixx("imagename", VNC["server"], 9)
	p5ixx("services", "repeater_service", 9)
	if proxy {
		p5ixx("services", "repeater_service", 5)
	}
	publicURL, forwardsTo, errC = ngrokAPI(NGROK_API_KEY)
	if !first && (strings.HasPrefix(forwardsTo, ifs) || errC != nil) && vh(args) {
		letf.Println("loop detected")
		return
	}
	PrintOk(forwardsTo, errC)

	ip = strings.Split(ips[len(ips)-1], "/")[0]
	if errC == nil {
		tcp, err = url.Parse(publicURL)
		if err != nil {
			letf.Println(err)
			return
		}
		connect, listen, inLAN = fromNgrok(forwardsTo)
		letf.Println(connect, listen, inLAN)
		if rProxy {
			plusS := "+"
			if rProxy2 {
				PrintOk("Is VNC proxyII listen - VNC проксиII ожидает подключения?", errC)
				//192.168.0.2->19216802
				id = "00000" + strings.ReplaceAll(ip, ".", "")
				id = strings.Trim(id[len(id)-9:], "0")
				plusS += id
			} else {
				PrintOk("Is VNC proxy listen - VNC прокси ожидает подключения?", errC)
			}
			if len(args) == 1 {
				if inLAN != "" && first {
					args = append(args, plusS) //server
				} else {
					args = append(args, ":") //viewer
				}
			}
		} else {
			if RportRFB != "" {
				PrintOk("Is VNC server listen - VNC экран ожидает подключения?", errC)
				if len(args) == 1 {
					if inLAN != "" {
						args = append(args, listen) //viewer
					} else {
						args = append(args, ":") //viewer
					}
				}
			}
			if RportViewer > 0 {
				PrintOk("Is VNC viewer listen - VNC наблюдатель ожидает подключения?", errC)
			}
		}
	}
	if len(args) > 1 {
		plus = strings.HasPrefix(args[1], "+")
		i, err := strconv.Atoi(args[1])
		if err != nil {
			switch {
			case args[1] == "-":
				// - try connect server to viewer via ngrok (revers)
				serverNgrok(args...)
			case strings.HasPrefix(args[1], "-"):
				// - try connect server to viewer via LAN (revers)
				serverLAN(args...)
			case strings.HasPrefix(args[1], "::") || plus:
				// :: as ::5900
				server(args...)
			default:
				// host[::port] [password] as LAN viewer connect mode
				// host[:display] [password] as LAN viewer connect mode
				// :host[::port] [password] as ngrok~proxy~LAN viewer connect mode
				// :host[:display] [password] as ngrok~proxy~LAN viewer connect mode
				// :id[:123456789] [password] as ngrok~proxy~IP viewer connect mode
				// : [password] as ngrok viewer connect mode
				// host as host:0 as host: as host::5900 as host::
				proxy = false
				proxy2 = false
				viewer(args...)
			}
			return
		}
		// -port as LAN viewer listen mode
		// port as ngrok viewer listen mode
		// 0  as 5500
		//+id as server with proxy  -id:id
		if plus {
			plus2 = true
			id = strconv.Itoa(i)
			server(args...)
		} else {
			proxy = false
			proxy2 = false
			viewerl(args...)
		}
		return
	}
	// as GUI or reg RfbPort
	server(args...)

}

func abs(s string) string {
	if strings.HasPrefix(s, "-") || (plus && !rProxy) {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		return s[1:]
	}
	return s
}
func usage() {
	li.Println("Usage - использование 8><--------------------------------------------")
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
	li.Println("--------------------------------------------><8")
}

func vh(args []string) bool {
	if len(args) < 2 {
		return true
	}
	parts := strings.Split(args[1], ".")
	if len(parts) != 4 {
		return true
	}
	parts[0] = strings.TrimPrefix(parts[0], ":")
	parts[3] = strings.Split(parts[3], ":")[0]
	return strings.Contains(ifs, strings.Join(parts, "."))
}
