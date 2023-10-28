package main

import (
	"strconv"

	"github.com/xlab/closer"
)

func serverLAN(args ...string) {
	var (
		port string
		diff bool
	)
	defer closer.Close()
	closer.Bind(cleanupS)

	ltf.Println(args)
	li.Printf("\"%s\" -host\n", args[0])

	if len(args) > 1 {
		tcp = abs(args[1])
		if tcp == ":" {
			tcp = ""
		}
	}
	tcp, port, diff = hp(tcp, strconv.Itoa(portViewer))
	li.Println("host", tcp)

	li.Println("On the other side was launched - на другой стороне был запушен")
	p, er := strconv.Atoi(port)
	if !diff || er != nil {
		p = CportViewer
	}
	li.Printf("`ngrokVNC -%d [password]`\n", p-CportViewer)
	li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via LAN - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через LAN")
	li.Println("The VNC server connects to the waiting VNC viewer via LAN - экран VNC подключается к ожидающему VNC наблюдателю через LAN")
	ll()
	err = serv()
	if err != nil {
		err = srcError(err)
		return
	}
	watch(false)
}
