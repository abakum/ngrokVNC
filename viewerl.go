package main

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"golang.org/x/sys/windows/registry"
)

func viewerl(args ...string) {
	ltf.Println(args)
	li.Printf("\"%s\" [-]port [password]\n", args[0])
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
		opts = options(append(opts, port))
	default:
		opts = append(opts, port)
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
