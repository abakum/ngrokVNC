package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xlab/closer"
)

func viewer() {
	var (
		err       error
		tvnviewer = filepath.Join(TightVNC, "tvnviewer.exe")
		host,
		publicURL,
		pass string
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

	li.Println("Viewer mode - наблюдатель подключается к ожидающему экрану VNC")
	li.Println(os.Args[0], "[host] [password]")
	li.Println(os.Args)

	// host[::port] [password] as LAN viewer connect mode
	// host[:screen] [password] as LAN viewer connect mode
	// : [password] as ngrok viewer connect mode
	// host as host:0 as host::5900
	if len(os.Args) > 1 {
		host = abs(os.Args[1])
		if host == ":" {
			host = ""
		}
	}

	if len(os.Args) > 2 {
		pass = "-password=" + os.Args[2]
	}

	if host != "" || NGROK_API_KEY == "" {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		li.Println("LAN mode - режим локальной сети")
		switch {
		case strings.HasSuffix(host, "::"):
			host += port
		case !strings.Contains(host, "::"):
			host += "::" + port
		}
	} else {
		li.Println("ngrok mode - режим ngrok туннеля")
		publicURL, _, err = ngrokAPI(NGROK_API_KEY)
		if err != nil {
			return
		}

		tcp, err = url.Parse(publicURL)
		if err != nil {
			err = srcError(err)
			return
		}
		host = strings.Replace(tcp.Host, ":", "::", 1)
	}
	li.Println("host", host)

	viewer := exec.Command(
		tvnviewer,
		host,
		pass,
	)
	viewer.Stdout = os.Stdout
	viewer.Stderr = os.Stderr

	closer.Bind(func() {
		if viewer.Process != nil && viewer.ProcessState == nil {
			PrintOk(fmt.Sprint("Kill ", viewer.Args), viewer.Process.Kill())
		}
	})
	li.Println(viewer.Args)
	err = viewer.Run()
}
