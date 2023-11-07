// try test ngrok.ListenAndForward
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/sync/errgroup"
)

// ListenAndForward creates a new [Forwarder] after connecting a new [Session], and
// then forwards all connections to the provided URL.
// This is a shortcut for calling [Connect] then [Session].ListenAndForward.
//
// Access to the underlying [Session] that was started automatically can be
// accessed via [Forwarder].Session.
//
// If an error is encountered during [Session].ListenAndForward, the [Session]
// object that was created will be closed automatically.
func ListenAndForward(ctx context.Context, backend *url.URL, tunnelConfig config.Tunnel, connectOpts ...ngrok.ConnectOption) (ngrok.Forwarder, error) {
	sess, err := ngrok.Connect(ctx, connectOpts...)
	if err != nil {
		return nil, err
	}
	fwd, err := ListenAndForwardSession(ctx, backend, tunnelConfig, sess)
	if err != nil {
		_ = sess.Close()
		return nil, err
	}

	return fwd, nil
}

func forwardTunnel(ctx context.Context, tun ngrok.Tunnel, url *url.URL) ngrok.Forwarder {
	mainGroup, ctx := errgroup.WithContext(ctx)
	fwdTasks := &sync.WaitGroup{}

	// sess := tun.Session()
	// sessImpl := sess.(*sessionImpl)
	// logger := sessImpl.inner().Logger.New("task", "forward", "toUrl", url, "tunnelUrl", tun.URL())
	ltf.Println("task", "forward", "toUrl", url, "tunnelUrl", tun.URL())

	mainGroup.Go(func() error {
		for {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}

			conn, err := tun.Accept()
			if err != nil {
				return err
			}
			fwdTasks.Add(1)

			go func() {
				ngrokConn := conn.(ngrok.Conn)
				// defer ngrokConn.Close()

				backend, err := openBackend(ctx, nil, tun, ngrokConn, url)
				if err != nil {
					defer ngrokConn.Close()
					// logger.Warn("failed to connect to backend url", "error", err)
					ltf.Println("failed to connect to backend url", "error", err)
					fwdTasks.Done()
					return
				}

				// defer backend.Close()
				join(ctx, ngrokConn, backend)
				fwdTasks.Done()
			}()
		}
	})

	return &forwarder{
		Tunnel:    tun,
		mainGroup: mainGroup,
	}
}

// TODO: use an actual reverse proxy for http/s tunnels so that the host header gets set?
func openBackend(ctx context.Context, logger log15.Logger, tun ngrok.Tunnel, tunnelConn ngrok.Conn, url *url.URL) (net.Conn, error) {
	host := url.Hostname()
	port := url.Port()
	if port == "" {
		switch {
		case usesTLS(url.Scheme):
			port = "443"
		case isHTTP(url.Scheme):
			port = "80"
		default:
			return nil, fmt.Errorf("no default tcp port available for %s", url.Scheme)
		}
		// logger.Debug("set default port", "port", port)
		ltf.Println("set default port", "port", port)
	}

	// Create TLS config if necessary
	var tlsConfig *tls.Config
	if usesTLS(url.Scheme) {
		tlsConfig = &tls.Config{
			ServerName:    url.Hostname(),
			Renegotiation: tls.RenegotiateOnceAsClient,
		}
	}

	dialer := &net.Dialer{}
	address := fmt.Sprintf("%s:%s", host, port)
	// logger.Debug("dial backend tcp", "address", address)
	ltf.Println("dial backend tcp", "address", address)

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		defer tunnelConn.Close()

		if isHTTP(tunnelConn.Proto()) {
			_ = writeHTTPError(tunnelConn, err)
		}
		return nil, err
	}

	if usesTLS(url.Scheme) && !tunnelConn.PassthroughTLS() {
		// logger.Debug("establishing TLS connection with backend")
		ltf.Println("establishing TLS connection with backend")

		return tls.Client(conn, tlsConfig), nil
	}

	return conn, nil
}

func usesTLS(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "https", "tls":
		return true
	default:
		return false
	}
}

func isHTTP(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "https", "http":
		return true
	default:
		return false
	}
}

func writeHTTPError(w io.Writer, err error) error {
	resp := &http.Response{}
	resp.StatusCode = http.StatusBadGateway
	resp.Body = io.NopCloser(bytes.NewBufferString(fmt.Sprintf("failed to connect to backend: %s", err.Error())))
	return resp.Write(w)
}

type forwarder struct {
	ngrok.Tunnel
	mainGroup *errgroup.Group
}

func (fwd *forwarder) Wait() error {
	return fwd.mainGroup.Wait()
}

type BindExtra struct {
	Token       string
	IPPolicyRef string
	Metadata    string
}

type tunnelConfigPrivate interface {
	ForwardsTo() string
	Extra() BindExtra
	Proto() string
	Opts() any
	Labels() map[string]string
	WithForwardsTo(string)
}

// %GOMODCACHE%\golang.ngrok.com\ngrok@v1.5.1\session.go
// c:\Users\kga\go\pkg\mod\golang.ngrok.com\ngrok@v1.5.1\session.go
// C:\Users\user\go\pkg\mod\golang.ngrok.com\ngrok@v1.5.1\session.go
func ListenAndForwardSession(ctx context.Context, url *url.URL, cfg config.Tunnel, s ngrok.Session) (ngrok.Forwarder, error) {
	tunnelCfg, ok := cfg.(tunnelConfigPrivate)
	if !ok {
		return nil, errors.New("invalid tunnel config")
	}

	// Set 'Forwards To'
	if tunnelCfg.ForwardsTo() == "" {
		tunnelCfg.WithForwardsTo(url.Host)
	}

	tun, err := s.Listen(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return forwardTunnel(ctx, tun, url), nil
}

// %GOMODCACHE%\golang.ngrok.com\ngrok@v1.5.1\forward.go
// c:\Users\kga\go\pkg\mod\golang.ngrok.com\ngrok@v1.5.1\forward.go
// C:\Users\user\go\pkg\mod\golang.ngrok.com\ngrok@v1.5.1\forward.go
func join_(ctx context.Context, left, right net.Conn) {
	g := &sync.WaitGroup{}
	g.Add(2)
	go func() {
		_, _ = io.Copy(left, right)
		left.Close()
		g.Done()
	}()
	go func() {
		_, _ = io.Copy(right, left)
		right.Close()
		g.Done()
	}()
	g.Wait()
}

func join(ctx context.Context, left, right net.Conn) {
	g, _ := errgroup.WithContext(ctx) // when ctx is canceled (on WithStopHandler or WithDisconnectHandler ) interrupts both io.Copy
	g.Go(func() error {
		_, err := io.Copy(left, right)
		left.Close() // on left disconnection interrupts io.Copy(right, left)
		return err
	})
	g.Go(func() error {
		_, err := io.Copy(right, left)
		right.Close() // on right disconnection interrupts io.Copy(left, right)
		return err
	})
	g.Wait()

}
