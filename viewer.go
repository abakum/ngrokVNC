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
	var (
		host string
	)
	defer closer.Close()
	closer.Bind(cleanup)

	ltf.Println(args)
	li.Printf("\"%s\" {[:]host[::port]|[:]host[:display]|:id[:123456789]|:} [password]\n", args[0])
	li.Println("On the other side may have been running - на другой стороне возможно был запущен")
	switch {
	case RportRFB == CportRFB:
		li.Println("\t`ngrokVNC`")
	case RportRFB != "":
		li.Printf("\t`ngrokVNC ::%s`\n", RportRFB)
	default:
		li.Println("\t`ngrokVNC [::port]`")
	}

	if len(args) > 1 {
		// host::port [password] as LAN viewer connect mode no crypt
		// host[:display] [password] as LAN viewer connect mode crypt
		// host as host: as host:0 as host:: as host::5900
		host = abs(args[1])
		if strings.HasPrefix(host, ":") {
			// :host::port [password] as ngrok~proxy~IP viewer connect mode no crypt
			// :host[:display] [password] as ngrok~proxy~IP viewer connect mode
			// :id [password] as ngrok~proxy~ID viewer connect mode
			// : [password] as ngrok viewer connect mode
			proxy = host != ":" && errNgrokAPI == nil && VNC["name"] == "UltraVNC"
			host = strings.TrimPrefix(host, ":")
		}
	}

	via := []string{"LAN", "LAN"}
	LAN := host != "" ||
		// emulate LAN mode?
		NGROK_AUTHTOKEN == ""
	if LAN {
		if !proxy {
			if strings.Contains(host, "::") {
				UseDSMPlugin = "0"
			}
		}
		h, _, _ := hp(host, portRFB)
		opts = append(opts, h)
	}
	if proxy || !LAN {
		via = []string{"ngrok", "туннель"}
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
		if tcp == "" {
			err = srcError(fmt.Errorf("no proxy to view - нет прокси для наблюдателя"))
		} else {
			opts = append(opts, tcp)
		}
	}
	li.Printf("The VNC viewer connects to the waiting VNC server via %s - наблюдатель VNC подключается к ожидающему экрану VNC через %s\n", via[0], via[1])

	if VNC["name"] == "UltraVNC" {
		opts = options(opts)
	}
	if len(opts) < 1 {
		err = srcError(fmt.Errorf("no host to view - нет адреса для наблюдателя"))
	}
	if err != nil {
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
	processName = VNC["viewer"]
	go watch(true)
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
				letf.Println("error write - ошибка записи", vnc)
			}
		}
	} else {
		opts = append(opts, "-noToolBar")
		if UseDSMPlugin == "1" && DSMPlugin != "" {
			opts = append(opts, "-DSMPlugin")
			opts = append(opts, DSMPlugin)
		}
		letf.Println("error read - ошибка чтения", vnc)
	}
	return
}
