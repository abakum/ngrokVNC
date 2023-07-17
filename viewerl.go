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

	li.Println("VNC viewer listen mode - VNC наблюдатель ожидает подключения экрана VNC")
	li.Println(os.Args[0], "port")
	li.Println(os.Args)

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
	}

	li.Println("port", p)
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
		planB(Errorf("empty NGROK_AUTHTOKEN"))
		return
	}

	_, forwardsTo, err := ngrokAPI(NGROK_API_KEY)
	if err == nil {
		planB(Errorf("found online client: %s", forwardsTo))
		return
	}
	err = nil

	err = run(context.Background(), fmt.Sprintf(":%d", p))

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err)
			err = nil
		}
	}
}
