// git clone github.com/abakum/ngrokVNC

// go get github.com/xlab/closer
// go get github.com/ngrok/ngrok-api-go/v5
// go get golang.ngrok.com/ngrok
// go get golang.org/x/sync/errgroup
// go get golang.org/x/sys/windows/registry
// go get gopkg.in/ini.v1
// go get github.com/lxn/win
// go get github.com/cakturk/go-netstat
// go get github.com/cakturk/go-netstat/netstat
// go install github.com/tc-hib/go-winres@latest
// go get github.com/mitchellh/go-ps
// go get github.com/zzl/go-win32api/v2

// go-winres init
// git tag v0.3.3-lw
// git push origin --tags

package main

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/ini.v1"
)

const (
	CportRFB    = "5900"
	CportViewer = 5500
	services    = "services.exe"
)

var (
	//go:embed NGROK_AUTHTOKEN.txt
	NGROK_AUTHTOKEN string
	//go:embed NGROK_API_KEY.txt
	NGROK_API_KEY string
	//go:embed uvnc.pkey.txt
	pkey []byte

	keyFN = "20230722_Viewer_ClientAuth.pkey"
	err,
	errNgrokAPI error
	TO          = time.Second * 60
	TOS         = time.Second * 7
	portRFB     = CportRFB
	portViewer  = CportViewer
	RportRFB    = ""
	RportViewer = 0
	PportRFB    = CportRFB
	PportViewer = CportViewer
	repeater    = "repeater.exe"
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
	processName,
	control,
	connect,
	publicURL,
	tcp,
	forwardsTo,
	listen,
	inLAN,
	ip,
	ifs,
	ultravnc,
	DSMPlugin,
	new string
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
	ips     []string
	servers int
	opts    = []string{}
	sRun,
	shutdown,
	cont,
	sConnect *exec.Cmd
	iniFile *ini.File
)

func main() {
	var (
		executable,
		cwd string
	)
	defer closer.Close()
	closer.Bind(cleanup)

	if len(os.Args) == 1 {
		usage()
	}
	executable, err = os.Executable()
	if err != nil {
		err = srcError(err)
		return
	}
	imagename := filepath.Base(executable)
	first = psCount(imagename, "") == 1

	ips = interfaces()
	if len(ips) == 0 {
		err = srcError(fmt.Errorf("not connected - нет сети"))
		return
	}
	ifs = strings.Join(ips, ",")

	cwd, err = os.Getwd()
	if err != nil {
		err = srcError(err)
		return
	}
	for _, xVNC := range VNCs {
		xVNC["path"] = filepath.Join(cwd, "..", xVNC["name"])
		if stat, err := os.Stat(xVNC["path"]); err == nil && stat.IsDir() {
			VNC = xVNC
			break
		}
	}

	key := `SOFTWARE\Classes\VncViewer.Config\DefaultIcon`
	k, er := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE|registry.SET_VALUE)
	if er == nil {
		key = ""
		old, _, er := k.GetStringValue(key)
		if er == nil {
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
			PrintOk(key, er)
		}
		k.Close()
	} else {
		PrintOk(key, er)
	}

	if VNC["path"] == "" {
		err = srcError(fmt.Errorf("not found VNC viewer - не найден VNC наблюдатель"))
		return
	}

	li.Println(VNC["name"], VNC["path"])
	processName = VNC["server"]
	serverExe = filepath.Join(VNC["path"], processName)
	viewerExe = filepath.Join(VNC["path"], VNC["viewer"])
	servers = psCount(processName, "")
	if VNC["name"] == "UltraVNC" {
		ultravnc = filepath.Join(VNC["path"], "ultravnc.ini")
		iniFile, err = ini.Load(ultravnc)
		if err != nil {
			err = srcError(err)
			return
		}
		section := iniFile.Section("admin")
		DSMPlugin = section.Key("DSMPlugin").String()
		if DSMPlugin != "" {
			UseDSMPlugin = "1"
		}
		keyFN = filepath.Join(VNC["path"], keyFN)
		_, er = os.Stat(keyFN)
		if er != nil {
			PrintOk(keyFN, os.WriteFile(keyFN, pkey, 0666))
		}
		hkl()
	}

	NGROK_AUTHTOKEN = Getenv("NGROK_AUTHTOKEN", NGROK_AUTHTOKEN) //create ngrok
	// NGROK_AUTHTOKEN += "-"                                       // emulate bad token or no internet
	// NGROK_AUTHTOKEN = ""                                         // emulate LAN mode
	NGROK_API_KEY = Getenv("NGROK_API_KEY", NGROK_API_KEY) //use ngrok

	args := os.Args[:]
	for k, v := range args {
		args[k] = strings.ReplaceAll(v, ";", ":")
	}

	p5ixx(9)
	processName = repeater
	p5ixx(9)
	if proxy {
		p5ixx(5)
	}
	processName = VNC["server"]

	publicURL, forwardsTo, errNgrokAPI = ngrokAPI(NGROK_API_KEY)
	if (strings.HasPrefix(forwardsTo, ifs) || errNgrokAPI != nil) && vh(args) {
		err = srcError(fmt.Errorf("loop detected - обнаружена петля"))
		return
	}
	PrintOk(forwardsTo, errNgrokAPI)

	ip = strings.Split(ips[len(ips)-1], "/")[0]
	if errNgrokAPI == nil {
		tcp = url2host(publicURL)
		connect, listen, inLAN = fromNgrok(forwardsTo)
		ltf.Println(connect, listen, inLAN)
		if rProxy {
			plusS := "+"
			if rProxy2 {
				PrintOk("Is VNC proxyII listen - VNC проксиII ожидает подключения?", errNgrokAPI)
				//192.168.0.2->19216802
				id = "00000" + strings.ReplaceAll(ip, ".", "")
				id = strings.Trim(id[len(id)-9:], "0")
				plusS += id
			} else {
				PrintOk("Is VNC proxy listen - VNC прокси ожидает подключения?", errNgrokAPI)
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
				PrintOk("Is VNC server listen - VNC экран ожидает подключения?", errNgrokAPI)
				if len(args) == 1 {
					if inLAN != "" {
						args = append(args, hpd(listen, RportRFB, CportRFB)) //viewer
					} else {
						args = append(args, ":") //viewer
					}
				}
			}
			if RportViewer > 0 {
				PrintOk("Is VNC viewer listen - VNC наблюдатель ожидает подключения?", errNgrokAPI)
				if len(args) == 1 {
					if inLAN != "" {
						args = append(args, fmt.Sprintf("-%s:%d", listen, RportViewer-CportViewer)) //serverLAN
					}
				}
			}
		}
	}
	if len(args) > 1 {
		plus = strings.HasPrefix(args[1], "+")
		i, er := strconv.Atoi(args[1])
		if er != nil {
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
				// host::port [password] as LAN viewer connect mode no crypt
				// host[:display] [password] as LAN viewer connect mode crypt
				// host as host: as host:0 as host:: as host::5900
				// :host::port [password] as ngrok~proxy~IP viewer connect mode no crypt
				// :host[:display] [password] as ngrok~proxy~IP viewer connect mode
				// :id [password] as ngrok~proxy~ID viewer connect mode
				// : [password] as ngrok viewer connect mode
				proxy = false
				proxy2 = false
				viewer(args...)
			}
			return
		}
		// -port as LAN viewer listen mode
		// port as ngrok viewer listen mode
		// 0  as 5500
		//+id as server with proxy -id:id
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
		NGROK_AUTHTOKEN = "" // not create ngrok
		NGROK_API_KEY = ""   // not use ngrok
		if strings.Count(s, ":") != 1 {
			UseDSMPlugin = "0"
		}
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
		return !first
	}
	parts := strings.Split(args[1], ".")
	if len(parts) != 4 {
		return !first
	}
	parts[0] = strings.TrimPrefix(parts[0], ":")
	parts[0] = strings.TrimPrefix(parts[0], "-")
	parts[3] = strings.Split(parts[3], ":")[0]
	return strings.Contains(ifs, strings.Join(parts, "."))
}

func url2host(publicURL string) string {
	tcp, err := url.Parse(publicURL)
	host := strings.Replace(publicURL, "tcp://", "", 1)
	if err == nil {
		host = tcp.Host
	}
	return strings.Replace(host, ":", "::", 1)
}

func cleanup() {
	if err != nil {
		let.Println(err)
		defer os.Exit(1)
	}
}
