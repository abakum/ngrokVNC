package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xlab/closer"
	"gopkg.in/ini.v1"
)

func viewer(args ...string) {
	ltf.Println("viewer", args)
	li.Printf("%q {[:]host[::port]|[:]host[:display]|:} [password]\n", args[0])
	li.Println("On the other side should be running - на другой стороне должен быть запущен")
	li.Println("`ngrokVNC [::port]`")
	var (
		err error
		host,
		publicURL string
		tcp *url.URL
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		// pressEnter()
	})

	// host[::port] [password] as LAN viewer connect mode
	// host[:display] [password] as LAN viewer connect mode
	// :host[::port] [password] as ngrok~proxy~LAN viewer connect mode
	// :host[:display] [password] as ngrok~proxy~LAN viewer connect mode
	// : [password] as ngrok viewer connect mode
	// host as host: as host:0 as host:: as host::5900
	if len(args) > 1 {
		host = abs(args[1])
		if strings.HasPrefix(host, ":") {
			proxy = host != ":" && NGROK_API_KEY != "" && VNC["name"] != "TightVNC"
			host = strings.TrimPrefix(host, ":")
		}
	}

	opts := []string{}
	via := []string{"LAN", "LAN"}
	LAN := host != "" || NGROK_API_KEY == ""
	if LAN {
		if !proxy {
			NGROK_AUTHTOKEN = "" // no ngrok
			NGROK_API_KEY = ""
		}
		h, _, _ := hp(host, portRFB)
		opts = append(opts, h)
	}
	if proxy || !LAN {
		via = []string{"ngrok", "туннель"}
		publicURL, _, err = ngrokAPI(NGROK_API_KEY)
		if err != nil {
			return
		}

		tcp, err = url.Parse(publicURL)
		if err != nil {
			err = srcError(err)
			return
		}
		if proxy {
			via = []string{"ngrok~proxy~LAN", "туннель~прокси~LAN"}
			switch VNC["name"] {
			case "UltraVNC":
				opts = append(opts, "-proxy")
			case "TurboVNC":
				opts = append(opts, "-via")
			}
		}
		opts = append(opts, strings.Replace(tcp.Host, ":", "::", 1))
	}
	li.Printf("The VNC viewer connects to the waiting VNC server via %s - наблюдатель VNC подключается к ожидающему экрану VNC через %s\n", via[0], via[1])

	if len(args) > 2 {
		switch VNC["name"] {
		case "TightVNC":
			opts = append(opts, "-password="+args[2])
		case "RealVNC":
		default:
			opts = append(opts, "-password")
			opts = append(opts, args[2])
		}
	}
	if VNC["name"] == "UltraVNC" {
		opts = append(opts, "-noToolBar")
		ultravnc := filepath.Join(VNC["path"], "ultravnc.ini")
		ini.PrettyFormat = false
		iniFile, err := ini.Load(ultravnc)
		DSMPlugin := ""
		UseDSMPlugin := "0"
		if err == nil {
			section := iniFile.Section("admin")
			DSMPlugin = section.Key("DSMPlugin").String()
			if section.Key("UseDSMPlugin").String() == "1" && DSMPlugin != "" {
				UseDSMPlugin = "1"
				opts = append(opts, "-DSMPlugin")
				opts = append(opts, DSMPlugin)
			}
		} else {
			letf.Println("error read", ultravnc)
		}
		ultravnc = filepath.Join(VNC["path"], "options.vnc")
		iniFile, err = ini.Load(ultravnc)
		if err == nil {
			section := iniFile.Section("options")
			if SetValue(section, "UseDSMPlugin", UseDSMPlugin) ||
				SetValue(section, "DSMPlugin", DSMPlugin) ||
				SetValue(section, "RequireEncryption", UseDSMPlugin) ||
				SetValue(section, "AllowUntrustedServers", "0") ||
				SetValue(section, "showtoolbar", "0") {
				err = iniFile.SaveTo(ultravnc)
				if err != nil {
					letf.Println("error write", ultravnc)
				}
			}
		} else {
			letf.Println("error read", ultravnc)
		}
	}
	viewer := exec.Command(viewerExe, opts...)
	viewer.Dir = filepath.Dir(viewer.Path)
	viewer.Stdout = os.Stdout
	viewer.Stderr = os.Stderr
	closer.Bind(func() {
		if viewer.Process != nil && viewer.ProcessState == nil {
			PrintOk(cmd("Kill", viewer), viewer.Process.Kill())
		}
	})
	li.Println(cmd("Run", viewer))
	err = viewer.Run()
}

func cmd(s string, c *exec.Cmd) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf(`%s "%s" %s`, s, c.Args[0], strings.Join(c.Args[1:], " "))
}
