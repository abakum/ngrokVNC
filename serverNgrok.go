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

func serverNgrok() {
	ltf.Println("serverNgrok", os.Args)
	li.Printf("%q -\n", os.Args[0])
	var (
		err error
		sRun,
		shutdown,
		cont,
		sConnect *exec.Cmd
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
	if NGROK_AUTHTOKEN == "" {
		err = fmt.Errorf("empty NGROK_AUTHTOKEN")
		return
	}

	li.Println("`ngrokVNC -`")
	li.Println("the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти")
	li.Println("\t`ngrokVNC 0`")
	for {
		publicURL, _, errC := ngrokAPI(NGROK_API_KEY)
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
		tcp, err := url.Parse(publicURL)
		host := publicURL
		if err == nil {
			host = strings.Replace(tcp.Host, ":", "::", 1)
		}
		sConnect := exec.Command(serverExe, append(arg,
			"-connect",
			host,
		)...)
		sConnect.Stdout = os.Stdout
		sConnect.Stderr = os.Stderr
		PrintOk(fmt.Sprint(sConnect.Args), sConnect.Run())
		for {
			new, _, errC := ngrokAPI(NGROK_API_KEY)
			remoteListen = errC == nil
			if !remoteListen || publicURL != new {
				PrintOk("VNC viewer connected - VNC наблюдатель подключен?", errC)
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
