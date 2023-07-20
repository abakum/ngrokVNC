package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows/registry"
)

func server() {
	ltf.Println("server", os.Args)
	li.Printf("%q [::port]\n", os.Args[0])
	li.Println("Run - запусти")
	li.Println("`ngrokVNC ::port`")
	li.Println("When there is no ngrok tunnel it will be created  - когда ngrok туннеля нет он создатся")
	li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC : [password]`")
	li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC host::port [password]`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC 0`")
	li.Println("This will create a ngrok tunnel - это создаст туннель")
	li.Println("The VNC viewer is waiting for the VNC server to connect via ngrok tunnel - наблюдатель VNC ожидает подключения VNC экрана через туннель")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC -0`")
	li.Println("The VNC viewer is waiting for the VNC server to be connected via LAN - наблюдатель VNC ожидает подключения VNC экрана через LAN")
	li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC -host`")
	li.Println()
	li.Println("Run - запусти")
	li.Println("`ngrokVNC -`")
	li.Println("the VNC server is waiting for ngrok tunnel of the VNC viewer to connect to it - экран VNC ожидает туннеля VNC наблюдателя чтоб к нему подключится")
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через ngrok на другой стороне запусти")
	li.Println("\t`ngrokVNC 0`")
	li.Println("`ngrokVNC -` unlike - в отличии от `ngrokVNC` does not stop working when the connection is broken - не прекращает работу при разрыве соединения")

	var (
		err error
	)
	defer closer.Close()

	closer.Bind(func() {
		if err != nil {
			let.Println(err)
			defer os.Exit(1)
		}
		// pressEnter()
	})

	// errD := dial(":" + port)
	// localListen := errD == nil
	// PrintOk("Is VNC service listen - экран VNC как сервис ожидает подключения наблюдателя?", errD)
	localListen := taskList("services eq tvnserver", "tvnserver")
	li.Println("Is VNC service listen - экран VNC как сервис ожидает подключения наблюдателя?", localListen)

	control := "-controlservice"
	k := registry.LOCAL_MACHINE
	if !localListen {
		control = "-controlapp"
		k = registry.CURRENT_USER
	}

	key := `SOFTWARE\TightVNC\Server`
	k, err = registry.OpenKey(k, key, registry.QUERY_VALUE|registry.SET_VALUE)
	if err == nil {
		key = "RfbPort"
		old, _, err := k.GetIntegerValue(key)
		if len(os.Args) > 1 {
			RfbPort, err := strconv.Atoi(strings.TrimPrefix(os.Args[1], "::"))
			if err != nil {
				RfbPort, _ = strconv.Atoi(port)
			}
			port = fmt.Sprintf("%d", RfbPort)
			if old != uint64(RfbPort) || err != nil {
				PrintOk(key, k.SetDWordValue(key, uint32(RfbPort)))
				if localListen {
					reload := exec.Command(
						tvnserver,
						control,
						"-reload",
					)
					PrintOk(fmt.Sprint(reload.Args), reload.Run())
				}
			}
		} else {
			if err == nil {
				port = fmt.Sprintf("%d", old)
			}
		}
		key = "AllowLoopback"
		old, _, err = k.GetIntegerValue(key)
		if old != 1 || err != nil {
			PrintOk(key, k.SetDWordValue(key, uint32(1)))
		}
		key = "LoopbackOnly"
		old, _, err = k.GetIntegerValue(key)
		if old != 0 || err != nil {
			PrintOk(key, k.SetDWordValue(key, uint32(0)))
		}
		k.Close()
	} else {
		PrintOk(key, err)
	}

	publicURL, _, errC := ngrokAPI(NGROK_API_KEY)
	remoteListen := errC == nil
	PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
	if !localListen {
		sRun := exec.Command(
			tvnserver,
			"-run",
		)
		sRun.Stdout = os.Stdout
		sRun.Stderr = os.Stderr
		closer.Bind(func() {
			if sRun.Process != nil && sRun.ProcessState == nil {
				shutdown := exec.Command(
					tvnserver,
					control,
					"-shutdown",
				)
				PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
			}
		})
		go func() {
			li.Println(sRun.Args)
			PrintOk(fmt.Sprint("Closed ", sRun.Args), sRun.Run())
			closer.Close()
		}()
		time.Sleep(time.Second)
	}

	cont := exec.Command(
		tvnserver,
		control,
	)
	cont.Stdout = os.Stdout
	cont.Stderr = os.Stderr
	closer.Bind(func() {
		if cont.Process != nil && cont.ProcessState == nil {
			PrintOk(fmt.Sprint("Kill ", cont.Args), cont.Process.Kill())
		}
	})
	go func() {
		li.Println(cont.Args)
		PrintOk(fmt.Sprint("Closed ", cont.Args), cont.Run())
		closer.Close()
	}()

	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"), ":"+port)
		return
	}

	if remoteListen {
		li.Println("On the other side was launched - на другой стороне был запушен")
		li.Println("`ngrokVNC 0`")
		li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via ngrok - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через туннель")
		li.Println("The VNC server connects to the waiting VNC viewer via ngrok - экран VNC подключается к ожидающему VNC наблюдателю через туннель")
		tcp, err := url.Parse(publicURL)
		host := publicURL
		if err == nil {
			host = strings.Replace(tcp.Host, ":", "::", 1)
		}
		sConnect := exec.Command(
			tvnserver,
			control,
			"-connect",
			host,
		)
		sConnect.Stdout = os.Stdout
		sConnect.Stderr = os.Stderr
		if !localListen {
			closer.Bind(func() {
				if sConnect.Process != nil && sConnect.ProcessState == nil {
					shutdown := exec.Command(
						tvnserver,
						control,
						"-shutdown",
					)
					PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
				}
			})
		}
		PrintOk(fmt.Sprint(sConnect.Args), sConnect.Run())
		closer.Hold()
		return
	}

	li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
	li.Println("\ton TCP port", port)
	li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
	li.Println("\t`ngrokVNC : [password]`")
	li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
	li.Println("\t`ngrokVNC host::port [password]`")
	err = run(context.Background(), ":"+port)

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, ":"+port)
			err = nil
		}
	}
}

func planB(err error, dest string) {
	s := "LAN mode - режим локальной сети"
	i := 0
	let.Println(err)
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, ifac := range ifaces {
			addrs, err := ifac.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ":") ||
					strings.HasPrefix(addr.String(), "127.") {
					continue
				}
				s += "\n\t" + addr.String()
				i++
			}
		}
	}
	if i > 0 {
		li.Println(s)
		for {
			time.Sleep(TO)
			if netstat("-a", dest) == "" {
				li.Println("no listen ", dest)
				break
			}
		}
		// closer.Close()
		// closer.Hold()
	} else {
		letf.Println("no ifaces for server")
	}
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string) error {
	ctxWT, caWT := context.WithTimeout(ctx, time.Second)
	defer caWT()
	sess, err := ngrok.Connect(ctxWT,
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
	)
	if err != nil {
		return Errorf("Connect %w", err)
	}
	sess.Close()

	ctx, ca := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			ca()
		}
	}()

	tun, err := ngrok.Listen(ctx,
		config.TCPEndpoint(),
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
		ngrok.WithStopHandler(func(ctx context.Context, sess ngrok.Session) error {
			go func() {
				time.Sleep(time.Millisecond * 10)
				ca()
			}()
			return nil
		}),
		ngrok.WithDisconnectHandler(func(ctx context.Context, sess ngrok.Session, err error) {
			PrintOk("WithDisconnectHandler", err)
			if err == nil {
				go func() {
					time.Sleep(time.Millisecond * 10)
					ca()
				}()
			}
		}),
	)
	if err != nil {
		return srcError(err)
	}

	ltf.Println("tunnel created:", tun.URL())

	for {
		conn, err := tun.Accept()
		if err != nil {
			return srcError(err)
		}

		ltf.Println("accepted connection from", conn.RemoteAddr())

		go PrintOk("connection closed:", handleConn(ctx, dest, conn))
		// go handleConn(ctx, dest, conn)
		if netstat("-a", dest) == "" {
			return srcError(fmt.Errorf("no listen %s", dest))
		}
	}
}

func handleConn(ctx context.Context, dest string, conn net.Conn) error {
	defer conn.Close()
	next, err := net.Dial("tcp", dest)
	if err != nil {
		return srcError(err)
	}
	defer next.Close()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := io.Copy(next, conn)
		next.(*net.TCPConn).CloseWrite()
		return srcError(err)
	})
	g.Go(func() error {
		_, err := io.Copy(conn, next)
		return srcError(err)
	})

	return g.Wait()
}

func dial(dest string) error {
	conn, err := net.Dial("tcp", dest)
	if err != nil {
		return srcError(err)
	}
	conn.Close()
	return err
}

func taskList(fi, ok string) bool {
	var (
		bBuffer bytes.Buffer
		err     error
	)
	list := exec.Command(
		"tasklist",
		"/nh",
		"/fi",
		fi,
	)
	list.Stdout = &bBuffer
	list.Stderr = &bBuffer
	err = list.Run()
	if err != nil {
		PrintOk(fmt.Sprint(list.Args), err)
		return false
	}
	return strings.Contains(bBuffer.String(), ok)
}
