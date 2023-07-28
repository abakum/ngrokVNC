package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
)

func serverLAN() {
	ltf.Println("serverLAN", os.Args)
	li.Printf("%q -host\n", os.Args[0])
	var (
		err error
		sRun,
		shutdown,
		cont,
		sConnect *exec.Cmd
		host,
		ESTABLISHED,
		new string
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		if sRun != nil {
			if sRun.Process != nil && sRun.ProcessState == nil && shutdown != nil {
				PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
				shutdown = nil
			}
		}
		if cont != nil {
			if cont.Process != nil && cont.ProcessState == nil {
				PrintOk(fmt.Sprint("Kill ", cont.Args), cont.Process.Kill())
			}
		}
		if sConnect != nil {
			if sConnect.Process != nil && sConnect.ProcessState == nil && shutdown != nil {
				PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
			}
		}
		// pressEnter()
	})

	if len(os.Args) > 1 {
		host = abs(os.Args[1])
		if host == ":" {
			host = ""
		}
	}
	host, _, _ = hp(host, strconv.Itoa(portViewer))
	hostD := strings.Replace(host, "::", ":", 1)
	li.Println("host", host)

	li.Println("On the other side was launched - на другой стороне был запушен")
	li.Println("`ngrokVNC -0`")
	li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via LAN - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через LAN")
	li.Println("The VNC server connects to the waiting VNC viewer via LAN - экран VNC подключается к ожидающему VNC наблюдателю через LAN")
	for {
		errC := dial(hostD)
		remoteListen := errC == nil
		if !remoteListen {
			time.Sleep(TO)
			continue
		}
		PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
		ll()
		arg := []string{}
		if VNC["name"] == "TightVNC" {
			arg = append(arg, control)
		}

		if !localListen {
			if shutdown == nil {
				shutdown = exec.Command(serverExe, append(arg, VNC["kill"])...)
			}
			if sRun == nil {
				sRun = exec.Command(serverExe,
					"-run",
				)
				sRun.Stdout = os.Stdout
				sRun.Stderr = os.Stderr
				go func() {
					li.Println(sRun.Args)
					PrintOk(fmt.Sprint("Closed ", sRun.Args), sRun.Run())
				}()
				time.Sleep(time.Second)
			}
		}

		if cont == nil {
			if VNC["name"] == "TightVNC" {
				cont = exec.Command(serverExe, arg...)
				cont.Stdout = os.Stdout
				cont.Stderr = os.Stderr
				go func() {
					li.Println(cont.Args)
					PrintOk(fmt.Sprint("Closed ", cont.Args), cont.Run())
					closer.Close()
				}()
			}
		}
		sConnect := exec.Command(serverExe, append(arg,
			"-connect",
			host,
		)...)
		sConnect.Stdout = os.Stdout
		sConnect.Stderr = os.Stderr
		PrintOk(fmt.Sprint(sConnect.Args), sConnect.Run())
		time.Sleep(time.Second)
		ESTABLISHED = netstat("", hostD, "")
		for {
			new = netstat("", hostD, "")
			if new == "" || new != ESTABLISHED {
				li.Println("VNC viewer connected - VNC наблюдатель подключен? no")
				break
			}
			time.Sleep(TO)
		}
		if shutdown != nil {
			PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
			shutdown = nil
		}
	}
}

func dial(dest string) error {
	conn, err := net.Dial("tcp", dest)
	if err != nil {
		return srcError(err)
	}
	conn.Close()
	return err
}

func hp(host, ps string) (hostPort, port string, ok bool) {
	switch {
	case strings.HasSuffix(host, "::"):
		host += ps
	case strings.Contains(host, "::"):
	case strings.HasSuffix(host, ":"):
		host += ":" + ps
	case strings.Contains(host, ":"):
		p, _ := strconv.Atoi(ps)
		parts := strings.Split(host, ":")
		i, err := strconv.Atoi(parts[1])
		if err == nil {
			i += p
		} else {
			i = p
		}
		host = fmt.Sprintf("%s::%d", parts[0], i)
	default:
		host += "::" + ps
	}
	port = strings.Split(host, "::")[1]
	return host, port, port != ps
}
