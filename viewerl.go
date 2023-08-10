package main

import (
	"context"
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

func viewerl(args ...string) {
	ltf.Println("viewerl", args)
	li.Printf(`"%s" [-]port\n`, args[0])
	var (
		err error
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		// pressEnter()
	})

	// -port as LAN viewer listen mode
	// port as ngrok viewer listen mode
	// 0  as 5500
	if len(args) > 1 {
		i, err := strconv.Atoi(abs(args[1]))
		if err == nil {
			if i < portViewer {
				portViewer += i
			} else {
				portViewer = i
			}
		}
		if strings.HasPrefix(args[1], "-") {
			li.Println("The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN")
			li.Println("\ton TCP port", portViewer)
			li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
			if i == 0 {
				li.Println("\t`ngrokVNC -host`")

			} else {
				li.Printf("\t`ngrokVNC -host:%d`", i)
			}
		} else {
			li.Println("This will create a ngrok tunnel - это создаст туннель")
			li.Println("The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через туннель")
			li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
			li.Println("\t`ngrokVNC`")
		}
	}
	OportViewer := portViewer
	portViewer = 0
	p5ixx("imagename", VNC["viewer"], 5)
	if portViewer > 0 {
		letf.Println(VNC["viewer"], "alredy listen on", portViewer)
		localListen = portViewer == OportViewer
		if !localListen {
			killViewer := exec.Command("taskKill", "/f", "/im", VNC["viewer"])
			killViewer.Stdout = os.Stdout
			killViewer.Stderr = os.Stderr
			PrintOk(cmd("Run", killViewer), killViewer.Run())
		}
	}
	portViewer = OportViewer

	opts := []string{"-listen"}
	port := strconv.Itoa(portViewer)

	switch VNC["name"] {
	case "TightVNC":
		// значение port появляется в поле `Accept Reverse connections on TCP port` на форме `TightVNC Viewer Configuration` но пока не кликнешь OK слушающий порт будет 5500
		key := `SOFTWARE\TightVNC\Viewer\Settings`
		k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE|registry.SET_VALUE)
		if err == nil {
			SetDWordValue(k, "ListenPort", portViewer)
			k.Close()
		} else {
			PrintOk(key, err)
		}
	case "UltraVNC":
		opts = append(opts, port)
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
				SetValue(section, "ListenPort", port) ||
				SetValue(section, "showtoolbar", "0") {
				err = iniFile.SaveTo(ultravnc)
				if err != nil {
					letf.Println("error write", ultravnc)
				}
			}
		} else {
			letf.Println("error read", ultravnc)
		}
	default:
		opts = append(opts, port)
	}

	if !localListen {
		viewer := exec.Command(viewerExe, opts...)
		viewer.Stdout = os.Stdout
		viewer.Stderr = os.Stderr
		closer.Bind(func() {
			if viewer.Process != nil && viewer.ProcessState == nil {
				PrintOk(cmd("Kill", viewer), viewer.Process.Kill())
			}
		})
		go func() {
			li.Println(cmd("Run", viewer))
			PrintOk(cmd("Closed", viewer), viewer.Run())
			if VNC["name"] != "TurboVNC" {
				closer.Close()
			}
		}()
		time.Sleep(time.Second)
	}
	port = ":" + port
	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"), port)
		return
	}

	if errC == nil {
		planB(Errorf("found online client: %s", forwardsTo), port)
		return
	}

	err = run(context.Background(), port, false)

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, port)
			err = nil
		}
	}
}
