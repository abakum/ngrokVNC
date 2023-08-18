package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xlab/closer"
)

func serverNgrok(args ...string) {
	ltf.Println("serverNgrok", args)
	li.Printf("\"%s\" -\n", args[0])
	var (
		err error
		sRun,
		shutdown,
		cont,
		sConnect *exec.Cmd
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
		// pressEnter()
	})
	if NGROK_AUTHTOKEN == "" {
		err = fmt.Errorf("empty NGROK_AUTHTOKEN")
		return
	}

	li.Println("`ngrokVNC -`")
	li.Println("the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти")
	li.Println("\t`ngrokVNC 0 [password]`")
	for {
		publicURL, _, errC = ngrokAPI(NGROK_API_KEY)
		remoteListen := errC == nil
		if !remoteListen {
			time.Sleep(TO)
			continue
		}
		PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
		ll()
		opts := []string{}
		if VNC["name"] == "TightVNC" {
			opts = append(opts, control)
		}

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
			if VNC["name"] == "TightVNC" {
				cont = exec.Command(serverExe, opts...)
				cont.Stdout = os.Stdout
				cont.Stderr = os.Stderr
				go func() {
					li.Println(cmd("Run", cont))
					PrintOk(cmd("Closed", cont), cont.Run())
					closer.Close()
				}()
			}
		}
		tcp, err := url.Parse(publicURL)
		host := strings.Replace(publicURL, "tcp://", "", 1)
		if err == nil {
			host = tcp.Host
		}
		host = strings.Replace(host, ":", "::", 1)
		if VNC["name"] == "UltraVNC" && localListen {
			setCommandLine(fmt.Sprintf("-autoreconnect -connect %s", host))
		} else {
			if VNC["name"] == "UltraVNC" {
				opts = append(opts, "-autoreconnect")
			}
			sConnect := exec.Command(serverExe, append(opts,
				"-connect",
				host,
			)...)
			sConnect.Stdout = os.Stdout
			sConnect.Stderr = os.Stderr
			PrintOk(cmd("Run", sConnect), sConnect.Run())
			time.Sleep(time.Second)
		}
		for {
			new, _, errC = ngrokAPI(NGROK_API_KEY)
			remoteListen = errC == nil
			if !remoteListen || publicURL != new {
				PrintOk("VNC viewer connected - VNC наблюдатель подключен?", errC)
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
