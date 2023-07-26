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
	ll()

	arg := []string{}

	switch VNC["name"] {
	case "TightVNC":
		arg = append(arg, control)

		key := `SOFTWARE\TightVNC\Server`
		k, err = registry.OpenKey(k, key, registry.QUERY_VALUE|registry.SET_VALUE)
		if err == nil {
			AcceptRfbConnections = GetBoolValue(k, "AcceptRfbConnections")
			key = "RfbPort"
			old, _, err := k.GetIntegerValue(key)
			if len(os.Args) > 1 {
				RfbPort, err := strconv.Atoi(strings.TrimPrefix(os.Args[1], "::"))
				if err != nil {
					RfbPort, _ = strconv.Atoi(portRFB)
				}
				portRFB = fmt.Sprintf("%d", RfbPort)
				if old != uint64(RfbPort) || err != nil {
					PrintOk(key, k.SetDWordValue(key, uint32(RfbPort)))
					if localListen {
						reload := exec.Command(
							serverExe,
							control,
							"-reload",
						)
						PrintOk(fmt.Sprint(reload.Args), reload.Run())
					}
				}
			} else {
				usage()
				if err == nil {
					portRFB = fmt.Sprintf("%d", old)
				}
				// p59("repeater")
			}
			SetDWordValue(k, "AllowLoopback", 1)
			SetDWordValue(k, "LoopbackOnly", 0)
			k.Close()
		} else {
			PrintOk(key, err)
		}
	case "UltraVNC":
		AcceptRfbConnections = true
	}
	p59xx(VNC["services"])
	p59xx("repeater_service")

	publicURL, _, errC := ngrokAPI(NGROK_API_KEY)
	remoteListen := errC == nil
	PrintOk("Is viewer listen - VNC наблюдатель ожидает подключения?", errC)
	if !localListen {
		sRun := exec.Command(serverExe,
			"-run",
		)
		sRun.Stdout = os.Stdout
		sRun.Stderr = os.Stderr
		closer.Bind(func() {
			if sRun.Process != nil && sRun.ProcessState == nil {
				shutdown := exec.Command(serverExe, append(arg, VNC["kill"])...)
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
	if VNC["name"] == "TightVNC" {
		cont := exec.Command(serverExe, arg...)
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
	}

	if NGROK_AUTHTOKEN == "" {
		planB(Errorf("empty NGROK_AUTHTOKEN"), ":"+portRFB)
		return
	}

	if remoteListen {
		li.Println("On the other side was launched - на другой стороне был запушен")
		li.Println("`ngrokVNC [-]port`")
		li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via ngrok - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через туннель")
		li.Println("The VNC server connects to the waiting VNC viewer via ngrok - экран VNC подключается к ожидающему VNC наблюдателю через туннель")
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
		if !localListen {
			closer.Bind(func() {
				if sConnect.Process != nil && sConnect.ProcessState == nil {
					shutdown := exec.Command(serverExe, append(arg, VNC["kill"])...)
					PrintOk(fmt.Sprint(shutdown.Args), shutdown.Run())
				}
			})
		}
		PrintOk(fmt.Sprint(sConnect.Args), sConnect.Run())
		closer.Hold()
		return
	}
	switch {
	case proxy:
		li.Println("The UltraVNC proxy is waiting for the UltraVNC viewer to connect -  UltraVNC прокси ожидает подключения UltraVNC наблюдателя")
		li.Println("\ton TCP port", portRFB)
		li.Println("\tTo view via ngrok~proxy~LAN on the other side, run - для просмотра через туннель~прокси~LAN на другой стороне запусти")
		li.Println("\t`ngrokVNC :host[::port] [password]`")
	case AcceptRfbConnections:
		li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
		li.Println("\ton TCP port", portRFB)
		li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
		li.Println("\t`ngrokVNC : [password]`")
		li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
		li.Println("\t`ngrokVNC host[::port] [password]`")
	}
	if AcceptRfbConnections {
		err = run(context.Background(), ":"+portRFB, false)
	}

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, ":"+portRFB)
			err = nil
		}
	}
}

func planB(err error, dest string) {
	if !AcceptRfbConnections {
		letf.Println("no accept connections")
		return
	}
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
		watch(dest)
	} else {
		letf.Println("no ifaces for server")
	}
}

func watch(dest string) {
	for {
		time.Sleep(TO)
		if netstat("-a", dest, "") == "" {
			li.Println("no listen ", dest)
			break
		}
	}
}

// https://github.com/ngrok/ngrok-go/blob/main/examples/ngrok-lite/main.go
func run(ctx context.Context, dest string, http bool) error {
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
	endpoint := config.TCPEndpoint()
	if http {
		endpoint = config.HTTPEndpoint()
	}
	tun, err := ngrok.Listen(ctx,
		endpoint,
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
	go func() {
		watch(dest)
		closer.Close()
	}()

	for {
		if netstat("-a", dest, "") == "" {
			return srcError(fmt.Errorf("no listen %s", dest))
		}
		conn, err := tun.Accept()
		if err != nil {
			return srcError(err)
		}

		ltf.Println("accepted connection from", conn.RemoteAddr())

		go PrintOk("connection closed:", handleConn(ctx, dest, conn))
		// go handleConn(ctx, dest, conn)
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

func taskList(fi string) string {
	var (
		bBuffer bytes.Buffer
	)
	list := exec.Command(
		"tasklist",
		"/nh",
		"/fi",
		fi,
	)
	list.Stdout = &bBuffer
	list.Stderr = &bBuffer
	err := list.Run()
	if err != nil {
		PrintOk(fmt.Sprint(list.Args), err)
		return ""
	}
	return bBuffer.String()
}

func GetBoolValue(k registry.Key, key string) bool {
	val, _, err := k.GetIntegerValue(key)
	if err == nil {
		return val == 1
	}
	return false
}

func SetDWordValue(k registry.Key, key string, val int) {
	old, _, err := k.GetIntegerValue(key)
	if old != uint64(val) || err != nil {
		PrintOk(key, k.SetDWordValue(key, uint32(val)))
	}
}

func netstat(a, host, pid string) (contains string) {
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
		if strings.Contains(contains, host) && strings.Contains(contains, ok) && strings.HasSuffix(contains, pid) {
			// ltf.Println(contains)
			return
		}
	}
}

func p59xx(serv string) {
	if serv == "" {
		return
	}
	parts := strings.Split(taskList("services eq "+serv), ".exe")
	if len(parts) < 2 {
		return
	}
	pid := strings.Split(strings.TrimSpace(parts[1]), " ")[0]
	pref := "  TCP    0.0.0.0:59"
	i, err := strconv.Atoi("59" + strings.Split(strings.TrimPrefix(netstat("-a", pref, pid), pref), " ")[0])
	if err != nil {
		return
	}
	if i > 5999 {
		return
	}
	proxy = serv == "repeater_service"
	portRFB = strconv.Itoa(i)
	ltf.Println(serv, portRFB)
}

func ll() {
	control = "-controlapp"
	k = registry.CURRENT_USER
	for _, xVNC := range VNCs {
		if xVNC["server"] == "" {
			continue
		}
		localListen = strings.Contains(taskList("services eq "+xVNC["services"]), xVNC["server"])
		if localListen {
			control = "-controlservice"
			k = registry.LOCAL_MACHINE
			VNC = xVNC
			break
		}
	}
	li.Println("Is VNC service listen - экран VNC как сервис ожидает подключения наблюдателя?", localListen, VNC["name"])
}
