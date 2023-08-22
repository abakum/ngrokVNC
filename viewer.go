package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xlab/closer"
	"gopkg.in/ini.v1"
)

func viewer(args ...string) {
	ltf.Println(args)
	li.Printf("\"%s\" {[:]host[::port]|[:]host[:display]|:id[:123456789]|:} [password]\n", args[0])
	li.Println("On the other side should be running - на другой стороне должен быть запущен")
	li.Println("`ngrokVNC [::port]`")
	var (
		err  error
		host string
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		// pressEnter()
	})

	// host[::port] [password] as LAN viewer connect mode no crypt
	// host[:display] [password] as LAN viewer connect mode
	// :host[::port] [password] as ngrok~proxy~IP viewer connect mode no crypt
	// :host[:display] [password] as ngrok~proxy~IP viewer connect mode
	// : [password] as ngrok viewer connect mode
	// host as host: as host:0 as host:: as host::5900
	// :id [password] as ngrok~proxy~ID viewer connect mode
	if len(args) > 1 {
		host = abs(args[1])
		if strings.HasPrefix(host, ":") {
			proxy = host != ":" && NGROK_API_KEY != "" && VNC["name"] != "TightVNC"
			host = strings.TrimPrefix(host, ":")
		}
	}

	opts := []string{}
	via := []string{"LAN", "LAN"}
	LAN := host != "" || NGROK_AUTHTOKEN == ""
	if LAN {
		if !proxy {
			NGROK_AUTHTOKEN = "" // no ngrok
			if strings.Contains(host, "::") {
				NGROK_API_KEY = "" // no crypt
			}
		}
		h, _, _ := hp(host, portRFB)
		opts = append(opts, h)
	}
	if proxy || !LAN {
		if proxy || rProxy {
			if rProxy2 {
				via = []string{"ngrok~proxy~ID", "туннель~прокси~ID"}
				if host == "" {
					opts = append(opts, "id:0")
				}
			} else {
				via = []string{"ngrok~proxy~IP", "туннель~прокси~IP"}
				if host == "" {
					opts = append(opts, listen)
				}
			}
			switch VNC["name"] {
			case "UltraVNC":
				opts = append(opts, "-proxy")
			case "TurboVNC":
				opts = append(opts, "-via")
			}
		}
		if tcp != nil {
			opts = append(opts, strings.Replace(tcp.Host, ":", "::", 1))
		} else {
			letf.Println("no proxy to view")
			return
		}
	}
	li.Printf("The VNC viewer connects to the waiting VNC server via %s - наблюдатель VNC подключается к ожидающему экрану VNC через %s\n", via[0], via[1])

	if VNC["name"] == "UltraVNC" {
		opts = options(opts)
	}
	if len(opts) < 1 {
		letf.Println("no host to view")
		return
	}
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

func options(o []string) (opts []string) {
	opts = o[:]
	if NGROK_API_KEY == "" {
		UseDSMPlugin = "0"
	}
	vnc := filepath.Join(VNC["path"], "options.vnc")
	ini.PrettyFormat = false
	iniFile, err := ini.Load(vnc)
	if err == nil {
		section := iniFile.Section("options")
		ok := SetValue(section, "UseDSMPlugin", UseDSMPlugin)
		ok = SetValue(section, "DSMPlugin", DSMPlugin) || ok
		ok = SetValue(section, "RequireEncryption", UseDSMPlugin) || ok
		ok = SetValue(section, "showtoolbar", "0") || ok
		if ok {
			err = iniFile.SaveTo(vnc)
			if err != nil {
				letf.Println("error write", vnc)
			}
		}
	} else {
		opts = append(opts, "-noToolBar")
		if UseDSMPlugin == "1" && DSMPlugin != "" {
			opts = append(opts, "-DSMPlugin")
			opts = append(opts, DSMPlugin)
		}
		letf.Println("error read", vnc)
	}
	return
}
