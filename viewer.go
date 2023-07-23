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
	li.Printf("%q {[:]host[::port]|[:]host[:screen]|:} [password]\n", os.Args[0])
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
	// host[:screen] [password] as LAN viewer connect mode
	// :host[::port] [password] as ngrok~proxy~LAN viewer connect mode
	// :host[:screen] [password] as ngrok~proxy~LAN viewer connect mode
	// : [password] as ngrok viewer connect mode
	// host as host:0 as host: as host::5900 as host::
	if len(os.Args) > 1 {
		host = abs(os.Args[1])
		if strings.HasPrefix(host, ":") {
			proxy = host != ":" && NGROK_API_KEY != "" && VNC["name"] != "TightVNC"
			host = strings.TrimPrefix(host, ":")
		}
	}

	arg := []string{}
	via := []string{"LAN", "LAN"}
	LAN := host != "" || NGROK_API_KEY == ""
	if LAN {
		if !proxy {
			NGROK_AUTHTOKEN = "" // no ngrok
			NGROK_API_KEY = ""
		}
		arg = append(arg, hp(host, portRFB))
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
			arg = append(arg, "-proxy")
		}
		arg = append(arg, strings.Replace(tcp.Host, ":", "::", 1))
	}
	li.Printf("The VNC viewer connects to the waiting VNC server via %s - наблюдатель VNC подключается к ожидающему экрану VNC через %s\n", via[0], via[1])

	if len(os.Args) > 2 {
		switch VNC["name"] {
		case "TightVNC":
			arg = append(arg, "-password="+os.Args[2])
		case "UltraVNC":
			arg = append(arg, "-password")
			arg = append(arg, os.Args[2])
		}
	}
	viewer := exec.Command(viewerExe, arg...)
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
