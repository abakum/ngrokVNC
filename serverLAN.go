package main

import (
	"bytes"
	"fmt"
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
		ps = fmt.Sprintf("%d", p)
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
	switch {
	case strings.HasSuffix(host, "::"):
		host += ps
	case !strings.Contains(host, "::"):
		host += "::" + ps
	case strings.HasSuffix(host, ":"):
		host += "::" + ps
	case strings.Contains(host, ":"):
		parts := strings.Split(host, ":")
		i, err := strconv.Atoi(parts[1])
		if err == nil {
			i += p
		} else {
			i = p
		}
		host = fmt.Sprintf("%s::%d", parts[0], i)
	}
	hostD := strings.Replace(host, "::", ":", 1)
	li.Println("host", host)

	li.Println("On the other side was launched - на другой стороне был запушен")
	li.Println("`ngrokVNC -0`")
	li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via LAN - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через LAN")
	li.Println("The VNC server connects to the waiting VNC viewer via LAN - экран VNC подключается к ожидающему VNC наблюдателю через туннель")
	for {
		errC := dial(hostD)
		remoteListen := errC == nil
		if !remoteListen {
			time.Sleep(TO)
			continue
		}
		PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
		errD := dial(":" + port)
		localListen := errD == nil
		PrintOk("Is VNC service listen - экран VNC как сервис ожидает подключения VNC наблюдателя?", errD)
		control := "-controlservice"
		if !localListen {
			control = "-controlapp"
			if shutdown == nil {
				shutdown = exec.Command(
					tvnserver,
					control,
					"-shutdown",
				)
			}
			if sRun == nil {
				sRun = exec.Command(
					tvnserver,
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
			cont = exec.Command(
				tvnserver,
				control,
			)
			cont.Stdout = os.Stdout
			cont.Stderr = os.Stderr
			go func() {
				li.Println(cont.Args)
				PrintOk(fmt.Sprint("Closed ", cont.Args), cont.Run())
				closer.Close()
			}()
		}
		sConnect = exec.Command(
			tvnserver,
			control,
			"-connect",
			host,
		)
		sConnect.Stdout = os.Stdout
		sConnect.Stderr = os.Stderr
		PrintOk(fmt.Sprint(sConnect.Args), sConnect.Run())
		time.Sleep(time.Second)
		ESTABLISHED = netstat("", hostD)
		for {
			new = netstat("", hostD)
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

func netstat(a, host string) (contains string) {
	var (
		bBuffer bytes.Buffer
		err     error
	)
	ok := "LISTENING"
	if a == "" {
		ok = "ESTABLISHED"
		a = "-o"
	}
	stat := exec.Command(
		"netstat",
		"-n",
		"-p",
		"TCP",
		"-o",
		a,
	)
	stat.Stdout = &bBuffer
	stat.Stderr = &bBuffer
	err = stat.Run()
	if err != nil {
		PrintOk(fmt.Sprint(stat.Args), err)
		return ""
	}

	for {
		contains, err = bBuffer.ReadString('\n')
		if err != nil {
			return ""
		}
		if strings.Contains(contains, host) && strings.Contains(contains, ok) {
			// ltf.Println(contains)
			return
		}
	}
}
