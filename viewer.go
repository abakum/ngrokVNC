package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/xlab/closer"
)

func viewer() {
	ltf.Println("viewer", os.Args)
	li.Printf("%q {host[::port]|host[:screen]|:} [password]\n", os.Args[0])
	li.Println("On the other side was launched - на другой стороне был запушен")
	li.Println("`ngrokVNC`")
	var (
		err error
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
	via := []string{"LAN", "LAN"}
	if host != "" || NGROK_API_KEY == "" {
		NGROK_AUTHTOKEN = "" // no ngrok
		NGROK_API_KEY = ""   // no crypt
		switch {
		case strings.HasSuffix(host, "::"):
			host += port
		case strings.Contains(host, ":"):
		case !strings.Contains(host, "::"):
			host += "::" + port
		}
	} else {
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
		host = strings.Replace(tcp.Host, ":", "::", 1)
	}
	li.Printf("The VNC viewer connects to the waiting VNC server via %s - наблюдатель VNC подключается к ожидающему экрану VNC через %s\n", via[0], via[1])

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
