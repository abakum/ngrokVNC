package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xlab/closer"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/ini.v1"
)

func server(args ...string) {
	defer closer.Close()
	closer.Bind(func() {
		setCommandLine("")
		cleanup()
	})

	ltf.Println(args)
	li.Printf("\"%s\" {+[id]|[::port]}\n", args[0])

	if len(args) > 1 {
		_, portRFB, reload = hp(abs(args[1]), portRFB)
	} else {
		reload = portRFB != CportRFB
		portRFB = CportRFB
	}
	ll()

	switch VNC["name"] {
	case "TightVNC":
		opts = append(opts, control)

		key := `SOFTWARE\TightVNC\Server`
		k, err = registry.OpenKey(k, key, registry.QUERY_VALUE|registry.SET_VALUE)
		if err != nil {
			err = srcError(err)
			return
		}
		AcceptRfbConnections = GetBoolValue(k, "AcceptRfbConnections")
		key = "RfbPort"
		old, _, er := k.GetIntegerValue(key)
		if reload {
			RfbPort, er := strconv.Atoi(portRFB)
			if old != uint64(RfbPort) || er != nil {
				PrintOk(key, k.SetDWordValue(key, uint32(RfbPort)))
				if localListen {
					reload := exec.Command(
						serverExe,
						control,
						"-reload",
					)
					PrintOk(cmd("Run", reload), reload.Run())
					p5ixx("imagename", VNC["server"], 9)
				}
			}
		} else {
			if er == nil {
				portRFB = fmt.Sprintf("%d", old)
			}
		}
		SetDWordValue(k, "AllowLoopback", 1)
		SetDWordValue(k, "LoopbackOnly", 0)
		k.Close()
	case "UltraVNC":
		AcceptRfbConnections = true
		ini.PrettyFormat = false
		iniFile, err = ini.LoadSources(ini.LoadOptions{
			IgnoreInlineComment: true,
		}, ultravnc)
		if err != nil {
			err = srcError(err)
			return
		}
		section := iniFile.Section("admin")
		AcceptRfbConnections = section.Key("SocketConnect").String() == "1"
		ok := SetValue(section, "PortNumber", portRFB)
		ok = SetValue(section, "UseDSMPlugin", UseDSMPlugin) || ok
		ok = SetValue(section, "AutoPortSelect", "0") || ok
		ok = SetValue(section, "AllowLoopback", "1") || ok
		ok = SetValue(section, "LoopbackOnly", "0") || ok
		if ok {
			err = iniFile.SaveTo(ultravnc)
			if err != nil {
				err = srcError(err)
				return
			}
		}
	}

	if VNC["name"] == "UltraVNC" {
		switch {
		case proxy2:
			connect = fmt.Sprintf("127.0.0.1::%d", PportViewer)
		case plus2:
		default:
			connect = ""
		}
	}
	if localListen {
		if connect == "" {
			err = setCommandLine("")
		} else {
			err = setCommandLine(fmt.Sprintf("-autoreconnect ID:%s -connect %s", id, connect))
		}
		if err != nil {
			err = srcError(err)
			return
		}
	} else {
		opts := []string{}
		if connect != "" {
			opts = append(opts,
				"-id:"+id,
				"-connect",
				connect,
			)
		} else {
			if servers > 0 {
				err = srcError(fmt.Errorf("VNC server already running - VNC экран уже запущен"))
				return
			}
		}
		sRun := exec.Command(serverExe, append(opts,
			"-run",
		)...)
		sRun.Stdout = os.Stdout
		sRun.Stderr = os.Stderr
		closer.Bind(func() {
			if sRun.Process != nil && sRun.ProcessState == nil {
				shutdown := exec.Command(serverExe, append(opts, VNC["kill"])...)
				PrintOk(cmd("Run", shutdown), shutdown.Run())
			}
		})
		go func() {
			li.Println(cmd("Run", sRun))
			PrintOk(cmd("Closed", sRun), sRun.Run())
			closer.Close()
		}()
		time.Sleep(time.Second)
	}
	if VNC["name"] == "TightVNC" {
		cont := exec.Command(serverExe, opts...)
		cont.Stdout = os.Stdout
		cont.Stderr = os.Stderr
		closer.Bind(func() {
			if cont.Process != nil && cont.ProcessState == nil {
				PrintOk(cmd("Kill", cont), cont.Process.Kill())
			}
		})
		go func() {
			li.Println(cmd("Run", cont))
			PrintOk(cmd("Closed", cont), cont.Run())
			closer.Close()
		}()
	}

	if NGROK_AUTHTOKEN == "" {
		li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
		li.Println("\ton TCP port", portRFB)
		li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
		li.Printf("\t`ngrokVNC %s [password]`", hpd(ip, portRFB, CportRFB))
		planB(Errorf("empty NGROK_AUTHTOKEN"), ":"+portRFB)
		return
	}

	if RportViewer > 0 && RportRFB == "" && tcp != "" { //&& !rProxy
		li.Println("On the other side was launched - на другой стороне был запушен")
		li.Printf("`ngrokVNC %d`", RportViewer-CportViewer)
		li.Println("On the other side the VNC viewer is waiting for the VNC server to be connected via ngrok - на другой стороне наблюдатель VNC ожидает подключения VNC экрана через туннель")
		li.Println("The VNC server connects to the waiting VNC viewer via ngrok - экран VNC подключается к ожидающему VNC наблюдателю через туннель")
		if localListen {
			err = setCommandLine(autoreconnect(tcp))
			if err != nil {
				err = srcError(err)
				return
			}
		} else {
			err = setCommandLine("")
			if err != nil {
				err = srcError(err)
				return
			}
			sConnect := exec.Command(serverExe, append(opts,
				"-connect",
				tcp,
			)...)
			sConnect.Stdout = os.Stdout
			sConnect.Stderr = os.Stderr
			closer.Bind(func() {
				if sConnect.Process != nil && sConnect.ProcessState == nil {
					shutdown := exec.Command(serverExe, append(opts, VNC["kill"])...)
					PrintOk(cmd("Run", shutdown), shutdown.Run())
				}
			})
			PrintOk(cmd("Run", sConnect), sConnect.Run())
			closer.Hold()
		}
	}
	switch {
	case proxy || plus || rProxy:
		li.Println("The UltraVNC proxy is waiting for the UltraVNC viewer to connect -  UltraVNC прокси ожидает подключения UltraVNC наблюдателя")
		li.Println("\ton TCP port", lPortRFB(RportRFB))
		if connect != "" {
			li.Println("The UltraVNC proxyII is waiting for the UltraVNC server to connect -  UltraVNC проксиII ожидает подключения UltraVNC экрана")
			li.Println("\ton TCP port", lPortViewer(RportViewer))
			li.Println("\tTo view via ngrok~proxy~ID on the other side, run - для просмотра через туннель~прокси~ID на другой стороне запусти")
			li.Printf("\t`ngrokVNC :%s [password]`", id)
		} else {
			li.Println("\tTo view via ngrok~proxy~IP on the other side, run - для просмотра через туннель~прокси~IP на другой стороне запусти")
			li.Printf("\t`ngrokVNC :%s [password]`", hpd(ip, portRFB, CportRFB))
			li.Println("\tTo view via LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
			li.Printf("\t`ngrokVNC %s [password]`", hpd(ip, portRFB, CportRFB))
		}
	case AcceptRfbConnections:
		li.Println("The VNC server is waiting for the VNC viewer to connect - экран VNC ожидает подключения VNC наблюдателя")
		li.Println("\ton TCP port", portRFB)
		li.Println("\tTo view via ngrok on the other side, run - для просмотра через туннель на другой стороне запусти")
		li.Println("\t`ngrokVNC : [password]`")
		li.Println("\tTo view via the LAN on the other side, run - для просмотра через LAN на другой стороне запусти")
		li.Printf("\t`ngrokVNC %s [password]`", hpd(ip, portRFB, CportRFB))
	}
	if plus {
		planB(fmt.Errorf("listen %s", portRFB), ":"+portRFB)
		err = nil
	}
	if AcceptRfbConnections {
		err = run(context.Background(), ":"+lPortRFB(portRFB), false)
	}

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err, ":"+portRFB)
			err = nil
		}
	}
}

func interfaces() (ifs []string) {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, ifac := range ifaces {
			addrs, err := ifac.Addrs()
			if err != nil || ifac.Flags&net.FlagUp == 0 || ifac.Flags&net.FlagRunning == 0 || ifac.Flags&net.FlagLoopback != 0 {
				continue
			}
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ":") {
					continue
				}
				ifs = append(ifs, addr.String())
			}
		}
	}
	return
}
func planB(err error, dest string) {
	if !AcceptRfbConnections {
		letf.Println("no accept connections - подключения запрещены")
		return
	}
	let.Println(err)
	li.Println("LAN mode - режим локальной сети")
	li.Println(ifs)
	watch(dest)
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
	endpoint := config.TCPEndpoint(config.WithForwardsTo(withForwardsTo(dest)))
	if http {
		endpoint = config.HTTPEndpoint(config.WithForwardsTo(withForwardsTo(dest)))
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

		ltf.Println("accepted connection from", conn.RemoteAddr(), "to", conn.LocalAddr())

		hkl(enHKL)

		go PrintOk("connection closed", handleConn(ctx, dest, conn))
	}
}
func hkl(enHKL uint32) {
	if VNC["name"] != "UltraVNC" {
		return
	}
	gkl := GetKeyboardLayout(0)
	if gkl != enHKL {
		li.Println("GetKeyboardLayout", gkl)
		PostMessage(HWND_BROADCAST, WM_INPUTLANGCHANGEREQUEST, 0, enHKL)
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
		next.(*net.TCPConn).CloseWrite() //for close without error
		time.Sleep(time.Millisecond * 7)
		next.Close()
		return srcError(err)
	})
	g.Go(func() error {
		_, err := io.Copy(conn, next)
		conn.Close()
		return srcError(err)
	})

	return g.Wait()
}

func taskList(fi string) string {
	bBuffer := tl(fi)
	return bBuffer.String()
}

func tl(fi string) (bBuffer bytes.Buffer) {
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
		PrintOk(cmd("Run", list), err)
		return
	}
	return
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
func ns(a string) string {
	var (
		err     error
		bBuffer bytes.Buffer
	)
	opts := []string{
		"-n",
		"-p",
		"TCP",
		"-o",
	}
	if a != "" {
		opts = append(opts, a)
	}
	stat := exec.Command("netstat", opts...)
	stat.Stdout = &bBuffer
	stat.Stderr = &bBuffer
	err = stat.Run()
	if err != nil {
		PrintOk(cmd("Run", stat), err)
		return ""
	}
	return bBuffer.String()
}

func nStat(all, a, host, pid string) (contains string) {
	var (
		err     error
		bBuffer bytes.Buffer
	)
	ok := "LISTENING"
	if a == "" {
		ok = "ESTABLISHED"
	}
	bBuffer.WriteString(all)
	for {
		contains, err = bBuffer.ReadString('\n')
		if err != nil {
			return ""
		}
		if strings.Contains(contains, host) && strings.Contains(contains, ok) && strings.Contains(contains, pid) {
			return
		}
	}
}

func netstat(a, host, pid string) (contains string) {
	return nStat(ns(a), a, host, pid)
}

func p5ixx(key, val string, i int) {
	if val == "" {
		return
	}
	s50 := strconv.Itoa(50 + i)
	bBuffer := tl(key + " eq " + val)
	all := ns("-a")
	for {
		line, err := bBuffer.ReadString('\n')
		if err != nil {
			return
		}
		parts := strings.Split(line, ".exe")
		if len(parts) < 2 {
			continue
		}
		pid := strings.Split(strings.TrimSpace(parts[1]), " ")[0]
		pref := "  TCP    0.0.0.0:" + s50
		suffix := strings.Split(strings.TrimPrefix(nStat(all, "-a", pref, pid), pref), " ")[0]
		if suffix == "" {
			continue
		}
		x, err := strconv.Atoi(s50 + suffix)
		if err != nil || x > (50+i+1)*100-1 || x < (50+i)*100 {
			continue
		}
		ltf.Println(key, val, x)
		if i == 9 {
			proxy = val == "repeater_service"
			if proxy {
				PportRFB = strconv.Itoa(x)
			} else {
				portRFB = strconv.Itoa(x)
			}
		} else {
			proxy2 = val == "repeater_service"
			if proxy2 {
				PportViewer = x
			} else {
				portViewer = x
			}
		}
	}
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

func SetValue(section *ini.Section, key, val string) (set bool) {
	set = section.Key(key).String() != val
	if set {
		ltf.Println(key, val)
		section.Key(key).SetValue(val)
	}
	return
}

func contains(net, ip string) bool {
	network, err := netip.ParsePrefix(net)
	if err != nil {
		return false
	}
	ipContains, err := netip.ParsePrefix(ip)
	if err != nil {
		return false
	}
	return network.Contains(ipContains.Addr())
}

func fromNgrok(forwardsTo string) (connect, listen, inLAN string) {
	netsPorts := strings.Split(forwardsTo, ":")
	nets := strings.Split(netsPorts[0], ",")
	for _, ip := range ips {
		for _, net := range nets {
			listen = strings.Split(net, "/")[0]
			if !contains(net, ip) {
				continue
			}
			inLAN = listen
		}
	}
	ltf.Println(netsPorts, listen, inLAN)
	if len(netsPorts) > 1 {
		if strings.HasPrefix(netsPorts[1], "59") {
			RportRFB = netsPorts[1]
		}
		if strings.HasPrefix(netsPorts[1], "55") {
			RportViewer, _ = strconv.Atoi(netsPorts[1])
		}
		if RportViewer > 0 {
			//case listen then ignore proxy
			return
		}
	}
	if len(netsPorts) > 2 {
		rProxy = true
		if netsPorts[2] == "" {
			RportViewer = 0
			return
		}
		rProxy2 = true
		RportViewer, _ = strconv.Atoi(netsPorts[2])
		if inLAN != "" {
			if RportViewer == CportViewer {
				connect = inLAN
				return
			}
			connect = fmt.Sprintf("%s::%d", inLAN, RportViewer)
		}
	}
	return
}
func lPortRFB(port string) string {
	if proxy {
		return PportRFB
	}
	return port
}
func lPortViewer(port int) int {
	if proxy2 {
		return PportViewer
	}
	return port
}
func hpd(h, p, c string) string {
	if UseDSMPlugin == "0" {
		if p == c {
			return h + "::"
		}
		return h + "::" + p
	} else {
		if p == c {
			return h
		}
		Ip, er := strconv.Atoi(p)
		Cp, _ := strconv.Atoi(c)
		if er != nil {
			Ip = Cp
		}
		return fmt.Sprintf("%s:%d", h, Ip-Cp)
	}
}

func withForwardsTo(lPort string) (meta string) {
	meta = ifs + lPort
	if proxy {
		meta += ":"
	}
	if proxy2 {
		meta += strconv.Itoa(PportViewer)
	}
	ltf.Println("withForwardsTo", meta)
	return
}

func setCommandLine(serviceCommandLine string) (err error) {
	if ultravnc == "" {
		return
	}
	ini.PrettyFormat = false
	iniFile, err = ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, ultravnc)
	if err != nil {
		return
	}
	section := iniFile.Section("admin")
	ok := reload
	ok = SetValue(section, "service_commandline", serviceCommandLine) || ok
	ok = SetValue(section, "UseDSMPlugin", UseDSMPlugin) || ok
	if ok {
		err = iniFile.SaveTo(ultravnc)
		if err != nil {
			return
		}
		if localListen {
			stop := exec.Command(
				"cmd",
				"/c",
				fmt.Sprintf("net stop %s&%s -startservice", VNC["services"], filepath.Base(serverExe)))
			stop.Dir = filepath.Dir(serverExe)
			stop.Stdout = os.Stdout
			stop.Stderr = os.Stderr
			PrintOk(cmd("Run", stop), stop.Run())
			time.Sleep(time.Second)
		}
		p5ixx("imagename", VNC["server"], 9)
	}
	return
}

func autoreconnect(tcp string) (a string) {
	a = "-autoreconnect"
	if VNC["name"] == "UltraVNC" {
		opts = append(opts, a)
		a += " "
	} else {
		a = ""
	}
	a += "-connect " + tcp
	return
}
