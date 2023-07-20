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
		i, _ := strconv.Atoi(abs(os.Args[1]))
		if i < p {
			p += i
		} else {
			p = i
		}
		if strings.HasPrefix(os.Args[1], "-") {
			li.Println("The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN")
			li.Println("\ton TCP port", p)
			li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
			li.Println("\t`ngrokVNC -host`")
		} else {
			li.Println("This will create a ngrok tunnel - это создаст туннель")
			li.Println("The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через туннель")
			li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
			li.Println("\t`ngrokVNC`")
		}
	}

	port := fmt.Sprintf(":%d", p)

	// значение port появляется в поле `Accept Reverse connections on TCP port` на форме `TightVNC Viewer Configuration` но пока не кликнешь OK слушающий порт будет 5500
	key := `SOFTWARE\TightVNC\Viewer\Settings`
	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE|registry.SET_VALUE)
	PrintOk(key, err)
	if err == nil {
		key = "ListenPort"
		old, _, err := k.GetIntegerValue(key)
		if old != uint64(p) || err != nil {
			PrintOk(key, k.SetDWordValue(key, uint32(p)))
		}
		k.Close()
	}

	cmd := "-listen"
	viewer := exec.Command(
		tvnviewer,
		cmd,
	)
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
		closer.Close()
	}()
	time.Sleep(time.Second)

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

	err = run(context.Background(), port)

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, port)
			err = nil
		}
	}
}
