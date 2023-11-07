package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	// "github.com/cakturk/go-netstat/netstat"
	"github.com/abakum/go-netstat/netstat"
	"github.com/mitchellh/go-ps"
	"github.com/xlab/closer"
	"github.com/zzl/go-win32api/v2/win32"

	// "github.com/zzl/go-win32api/v2/win32"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	ngrok_log "golang.ngrok.com/ngrok/log"
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
					p5ixx(9)
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
		planB(Errorf("empty NGROK_AUTHTOKEN"))
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
		planB(fmt.Errorf("listen %s", portRFB))
		err = nil
	}
	if AcceptRfbConnections {
		// err = run(context.Background(), ":"+lPortRFB(portRFB), false)
		err = run2(context.Background(), ":"+lPortRFB(portRFB), false)
	}

	if err != nil {
		if strings.Contains(err.Error(), "ERR_NGROK_105") ||
			strings.Contains(err.Error(), "failed to dial ngrok server") {
			planB(err)
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
func planB(err error) {
	if !AcceptRfbConnections {
		letf.Println("no accept connections - подключения запрещены")
		return
	}
	let.Println(err)
	li.Println("LAN mode - режим локальной сети")
	li.Println(ifs)
	watch(false)
}

// break or closer.Close() on `Stopped TCP`,
// change input language on `Disconnect TCP` or `Changed TCP`
func watch(close bool) {
	old := -1
	ste_ := ""
	for {
		time.Sleep(TOS)
		ste := ""
		new := netSt(func(s *netstat.SockTabEntry) bool {
			ok := s.Process != nil && s.Process.Name == processName && (s.State == netstat.Listen || s.State == netstat.Established)
			if ok {
				ste += fmt.Sprintln("\t", s.LocalAddr, s.RemoteAddr, s.State)
			}
			return ok
		})
		if new == 0 {
			lt.Println("Stopped TCP")
			if close {
				closer.Close()
			}
			break
		}
		if old != new {
			if old > new {
				lt.Print("Disconnect TCP\n", ste)
				hkl()
			} else {
				if strings.Contains(ste, "ESTABLISHED") {
					lt.Print("Established TCP\n", ste)
				} else {
					lt.Print("Listening TCP\n", ste)
				}
			}
			ste_ = ste
			old = new
		}
		if ste_ != ste {
			lt.Print("Changed TCP\n", ste)
			hkl()
			ste_ = ste
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
				time.Sleep(TOM)
				ca()
			}()
			return nil
		}),
		ngrok.WithDisconnectHandler(func(ctx context.Context, sess ngrok.Session, err error) {
			PrintOk("WithDisconnectHandler", err)
			if err == nil {
				go func() {
					time.Sleep(TOM)
					ca()
				}()
			}
		}),
	)
	if err != nil {
		return srcError(err)
	}

	ltf.Println("tunnel created:", tun.URL())
	go watch(true)

	for {
		conn, err := tun.Accept()
		if err != nil {
			return srcError(err)
		}

		ltf.Println("accepted connection from", conn.RemoteAddr(), "to", conn.LocalAddr())

		go PrintOk("connection closed", handleConn(ctx, dest, conn))
	}
}

func run2(ctx context.Context, dest string, http bool) error {
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

	destURL, err := url.Parse("tcp://" + dest)
	if err != nil {
		return Errorf("Parse %w", err)
	}
	fwd, err := ngrok.ListenAndForward(ctx,
		destURL,
		endpoint,
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
		ngrok.WithStopHandler(func(ctx context.Context, sess ngrok.Session) error {
			go func() {
				time.Sleep(TOM)
				ca()
			}()
			return nil
		}),
		ngrok.WithDisconnectHandler(func(ctx context.Context, sess ngrok.Session, err error) {
			PrintOk("WithDisconnectHandler", err)
			if err == nil {
				go func() {
					time.Sleep(TOM)
					ca()
				}()
			}
		}),
		ngrok.WithLogger(&logger{lvl: ngrok_log.LogLevelDebug}),
	)
	if err != nil {
		return srcError(err)
	}

	ltf.Println("tunnel created:", fwd.URL())
	go watch(true)

	return srcError(fwd.Wait())
}

func hkl() {
	const (
		Tray   = "Shell_TrayWnd"
		usKLID = "00000409"
	)
	if VNC["name"] != "UltraVNC" {
		return
	}
	usHKL, er := win32.LoadKeyboardLayout(win32.StrToPwstr(usKLID), 0)
	if er != win32.NO_ERROR {
		letf.Println(er)
		return
	}
	// ltf.Println("LoadKeyboardLayout", usHKL)
	hwnd := win32.GetForegroundWindow()

	for {
		if hwnd == 0 {
			return
		}
		kl, class := gkl(hwnd)
		if kl == usHKL {
			return
		}
		ltf.Println("gkl", kl, class, hwnd)
		ret, er := win32.SendMessage(hwnd, win32.WM_INPUTLANGCHANGEREQUEST, 0, uintptr(usHKL))
		if ret != 0 || er != win32.NO_ERROR {
			ltf.Println("SendMessage", ret, er)
		}
		// 	ret, er := win32.PostMessage(hwnd, win32.WM_INPUTLANGCHANGEREQUEST, 0, uintptr(usHKL))
		// 	if ret != win32.TRUE || er != win32.NO_ERROR {
		// 		ltf.Println("PostMessage", ret, er)
		// 	}
		// 	time.Sleep(TOM)
		hwnd, er = win32.GetWindow(hwnd, win32.GW_HWNDPREV)
		if hwnd == 0 || er != win32.NO_ERROR {
			letf.Println(hwnd, er)
			continue
		}
		if class == Tray {
			win32.SetForegroundWindow(hwnd)
		}
	}
}

func BufToPwstr(size uint) *uint16 {
	buf := make([]uint16, size*2+1)
	return &buf[0]
}

func GetClassName(hwnd win32.HWND) (ClassName string) {
	const nMaxCount = 256

	if hwnd == 0 {
		return
	}

	// lpClassName := win32.StrToPwstr("")
	// lpClassName := win32.StrToPwstr(strings.Repeat(" ", nMaxCount))
	lpClassName := BufToPwstr(nMaxCount)
	copied, er := win32.GetClassName(hwnd, lpClassName, nMaxCount)
	if copied == 0 || er != win32.NO_ERROR {
		return
	}
	ClassName = win32.PwstrToStr(lpClassName)
	return
}

// func gkl(hwnd win.HWND) (kl uint32, class string) {
func gkl(hwnd win32.HWND) (kl unsafe.Pointer, class string) {
	const Console = "ConsoleWindowClass"
	var er win32.WIN32_ERROR
	if hwnd == 0 {
		return
	}
	class = GetClassName(hwnd)
	if class == Console {
		hwnd, er = win32.GetWindow(hwnd, win32.GW_HWNDPREV)
		if er != win32.NO_ERROR {
			letf.Println(er)
			return
		}
		if hwnd == 0 {
			return
		}
	}
	tid := win32.GetWindowThreadProcessId(hwnd, nil)
	kl = win32.GetKeyboardLayout(tid)
	return
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
		time.Sleep(TOM)
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

func psCount(name, parent string) (count int) {
	pes, err := ps.Processes()
	if err != nil {
		return
	}
	for _, p := range pes {
		if p == nil {
			continue
		}
		ok := true
		if parent != "" {
			pp, err := ps.FindProcess(p.PPid())
			if pp == nil || err != nil {
				continue
			}
			ok = pp.Executable() == parent
		}
		if p.Executable() == name && ok {
			count++
		}
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

// func(s *netstat.SockTabEntry) bool {return s.State == a}
func netSt(accept netstat.AcceptFn) int {
	tabs, err := netstat.TCPSocks(accept)
	if err != nil {
		return 0
	}
	return len(tabs)
}

func p5ixx(i int) {
	min := uint16((50 + i) * 100)
	max := min + 100
	tabs, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
		return s.State == netstat.Listen && s.LocalAddr.Port >= min && s.LocalAddr.Port < max && s.Process != nil && s.Process.Name == processName
	})
	if err != nil {
		return
	}
	for _, s := range tabs {
		x := int(s.LocalAddr.Port)
		ltf.Println(processName, x)
		if i == 9 {
			proxy = processName == repeater
			if proxy {
				PportRFB = strconv.Itoa(x)
			} else {
				portRFB = strconv.Itoa(x)
			}
		} else {
			proxy2 = processName == repeater
			if proxy2 {
				PportViewer = x
			} else {
				portViewer = x
			}
		}
		if true {
			return
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
		localListen = psCount(xVNC["server"], services) > 0
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
		p5ixx(9)
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

// Simple logger that forwards to the Go standard logger.
type logger struct {
	lvl ngrok_log.LogLevel
}

func (l *logger) Log(ctx context.Context, lvl ngrok_log.LogLevel, msg string, data map[string]interface{}) {
	if lvl > l.lvl {
		return
	}
	// lvlName, _ := ngrok_log.StringFromLogLevel(lvl)
	// log.Printf("[%s] %s %v", lvlName, msg, data)
	if msg != "heartbeat received" {
		ltf.Println(msg, data)
	}
}
