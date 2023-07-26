package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"golang.org/x/sys/windows/registry"
)

func viewerl() {
	ltf.Println("viewerl", os.Args)
	li.Printf("%q [-]port\n", os.Args[0])
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
	if len(os.Args) > 1 {
		i, err := strconv.Atoi(abs(os.Args[1]))
		if err == nil {
			if i < portViewer {
				portViewer += i
			} else {
				portViewer = i
			}
		}
		if strings.HasPrefix(os.Args[1], "-") {
			li.Println("The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN")
			li.Println("\ton TCP port", portViewer)
			li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
			li.Println("\t`ngrokVNC -host[::port]`")
		} else {
			li.Println("This will create a ngrok tunnel - это создаст туннель")
			li.Println("The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через туннель")
			li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
			li.Println("\t`ngrokVNC`")
		}
	}

	arg := []string{"-listen"}
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
	default:
		arg = append(arg, port)
	}

	if p55xx() != portViewer {
		viewer := exec.Command(viewerExe, arg...)
		viewer.Stdout = os.Stdout
		viewer.Stderr = os.Stderr
		closer.Bind(func() {
			if viewer.Process != nil && viewer.ProcessState == nil {
				PrintOk(fmt.Sprint("Kill ", viewer.Args), viewer.Process.Kill())
			}
		})
		go func() {
			li.Println(viewer.Args)
			PrintOk(fmt.Sprint("Closed ", viewer.Args), viewer.Run())
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

	_, forwardsTo, err := ngrokAPI(NGROK_API_KEY)
	if err == nil {
		planB(Errorf("found online client: %s", forwardsTo), port)
		return
	}
	err = nil

	err = run(context.Background(), port, false)

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, port)
			err = nil
		}
	}
}

func p55xx() (i int) {
	pref := "  TCP    0.0.0.0:55"
	ok := netstat("-a", pref, "")
	if ok == "" {
		return 0
	}
	parts := strings.Split(strings.TrimPrefix(ok, pref), " ")
	i, err := strconv.Atoi("55" + parts[0])
	if err != nil {
		return
	}
	if i > 5599 {
		return i
	}
	li.Println("listen", i)
	return i
}

func trimDubleSpace(s string) (trim string) {
	trim = s
	for strings.Contains(trim, "  ") {
		trim = strings.ReplaceAll(trim, "  ", " ")
	}
	return
}
