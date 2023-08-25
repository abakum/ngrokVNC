package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
)

func serverNgrok(args ...string) {
	defer closer.Close()
	closer.Bind(cleanupS)

	ltf.Println(args)
	li.Printf("\"%s\" -\n", args[0])

	if NGROK_API_KEY == "" {
		err = srcError(fmt.Errorf("empty NGROK_API_KEY - не задан NGROK_API_KEY"))
		return
	}

	li.Println("the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти")
	li.Println("\t`ngrokVNC 0 [password]`")
	for {
		publicURL, _, errNgrokAPI = ngrokAPI(NGROK_API_KEY)
		remoteListen := errNgrokAPI == nil
		if !remoteListen {
			time.Sleep(TO)
			continue
		}
		tcp = url2host(publicURL)
		PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errNgrokAPI)
		ll()
		opts = []string{}
		err = serv()
		if err != nil {
			err = srcError(err)
			return
		}
		for {
			new, _, errNgrokAPI = ngrokAPI(NGROK_API_KEY)
			remoteListen = errNgrokAPI == nil
			if !remoteListen || publicURL != new {
				PrintOk("VNC viewer connected - VNC наблюдатель подключен?", errNgrokAPI)
				break
			}
			time.Sleep(TO)
		}
		if shutdown != nil {
			PrintOk(cmd("Run", shutdown), shutdown.Run())
			shutdown = nil
		}
	}
}

func cleanupS() {
	if err != nil {
		let.Println(err)
		defer os.Exit(1)
	}
	if sRun != nil {
		if sRun.Process != nil && sRun.ProcessState == nil && shutdown != nil {
			PrintOk(cmd("Run", shutdown), shutdown.Run())
			shutdown = nil
		}
	}
	if cont != nil {
		if cont.Process != nil && cont.ProcessState == nil {
			PrintOk(cmd("Kill", cont), cont.Process.Kill())
		}
	}
	if sConnect != nil {
		if sConnect.Process != nil && sConnect.ProcessState == nil && shutdown != nil {
			PrintOk(cmd("Run", shutdown), shutdown.Run())
		}
	}
	setCommandLine("")
}

func serv() (err error) {
	if VNC["name"] == "TightVNC" {
		opts = append(opts, control)
		if !localListen {
			if shutdown == nil {
				shutdown = exec.Command(serverExe, append(opts, VNC["kill"])...)
			}
			if sRun == nil {
				sRun = exec.Command(serverExe,
					"-run",
				)
				sRun.Stdout = os.Stdout
				sRun.Stderr = os.Stderr
				go func() {
					li.Println(cmd("Run", sRun))
					PrintOk(cmd("Closed", sRun), sRun.Run())
				}()
				time.Sleep(time.Second)
			}
		}

		if cont == nil {
			cont = exec.Command(serverExe, opts...)
			cont.Stdout = os.Stdout
			cont.Stderr = os.Stderr
			go func() {
				li.Println(cmd("Run", cont))
				PrintOk(cmd("Closed", cont), cont.Run())
				closer.Close()
			}()
		}
		sConnect := exec.Command(serverExe, append(opts,
			"-connect",
			tcp,
		)...)
		sConnect.Stdout = os.Stdout
		sConnect.Stderr = os.Stderr
		PrintOk(cmd("Run", sConnect), sConnect.Run())
	} else {
		//VNC["name"] == "UltraVNC"
		if localListen {
			err = setCommandLine(autoreconnect(tcp))
			if err != nil {
				return
			}
		} else {
			if servers > 0 {
				err = srcError(fmt.Errorf("VNC server already running - VNC экран уже запущен"))
				return
			}
			err = setCommandLine("")
			if err != nil {
				return
			}
			if shutdown == nil {
				shutdown = exec.Command(serverExe, VNC["kill"])
			}
			sConnect := exec.Command(serverExe, append(opts,
				// "-autoreconnect",
				"-connect",
				tcp,
				"-run",
			)...)
			sConnect.Stdout = os.Stdout
			sConnect.Stderr = os.Stderr
			if !localListen {
				closer.Bind(func() {
					if sConnect.Process != nil && sConnect.ProcessState == nil {
						shutdown := exec.Command(serverExe, append(opts, VNC["kill"])...)
						PrintOk(cmd("Run", shutdown), shutdown.Run())
					}
				})
			}
			li.Println(cmd("Run", sConnect))
			PrintOk(cmd("Closed", sConnect), sConnect.Run())
		}
	}
	time.Sleep(time.Second)
	return
}

func hp(host, ps string) (hostPort, port string, diff bool) {
	switch {
	case strings.EqualFold("id", host):
		return host + ":0", ps, false
	case strings.HasSuffix(host, "::"):
		host += ps
	case strings.Contains(host, "::"):
	case strings.HasSuffix(host, ":"):
		host += ":" + ps
	case strings.Contains(host, ":"):
		p, _ := strconv.Atoi(ps)
		parts := strings.Split(host, ":")
		if strings.EqualFold("id", parts[0]) {
			return host, ps, false
		}
		i, err := strconv.Atoi(parts[1])
		if err == nil {
			i += p
		} else {
			i = p
		}
		host = fmt.Sprintf("%s::%d", parts[0], i)
	default:
		i, err := strconv.Atoi(host)
		if err == nil && i < 1000000000 && i > -1 {
			return "ID:" + host, ps, false
		}
		host += "::" + ps
	}
	port = strings.Split(host, "::")[1]
	return host, port, port != ps
}
