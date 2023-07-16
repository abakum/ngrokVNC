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

func serverc() {
	var (
		err error
		sRun,
		shutdown,
		cont,
		sConnect *exec.Cmd
		control = "-controlservice"
		TO      = time.Second * 10
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

	li.Println("VNC server try connect to viewer mode - экран VNC пытается подключается к ожидающему VNC наблюдателю")
	for {
		publicURL, _, errC := ngrokAPI(NGROK_API_KEY)
		rl := errC == nil
		PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
		if !rl {
			time.Sleep(TO)
			continue
		}
		errD := dial(":" + port)
		ll := errD == nil
		PrintOk("Is VNC service listen - экран VNC как сервис ожидает подключения VNC наблюдателя?", errD)
		control = "-controlservice"
		if !ll {
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
		tcp, err := url.Parse(publicURL)
		host := publicURL
		if err == nil {
			host = strings.Replace(tcp.Host, ":", "::", 1)
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
		for {
			_, _, errC = ngrokAPI(NGROK_API_KEY)
			rl = errC == nil
			PrintOk("VNC viewer connected - VNC наблюдатель подключен?", errC)
			if !rl {
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
