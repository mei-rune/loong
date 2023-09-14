package loong

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln TcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func RunTLS(network, addr, certFile, keyFile string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
}

func RunHTTP(network, addr string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	return http.ListenAndServe(addr, engine)
}

func RunServer(srv *http.Server, network, addr string, wrapFn ...func(net.Listener) net.Listener) error {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	tcpListener, ok := ln.(*net.TCPListener)
	if ok {
		ln = TcpKeepAliveListener{tcpListener}
	}

	if len(wrapFn) > 0 {
		ln = wrapFn[0](ln)
	}

	return srv.Serve(ln)
}

func RunServerTLS(srv *http.Server, network, addr, certFile, keyFile string, wrapFn ...func(net.Listener) net.Listener) error {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	tcpListener, ok := ln.(*net.TCPListener)
	if ok {
		ln = TcpKeepAliveListener{tcpListener}
	}

	if len(wrapFn) > 0 {
		ln = wrapFn[0](ln)
	}
	return srv.ServeTLS(ln, certFile, keyFile)
}

func IsSocketBindError(err error) bool {
	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}

func ListenAtDynamicPort(network, address string, portStart, portEnd int) (string, net.Listener, error) {
	// isHTTPs := false
	switch strings.ToLower(network) {
	case "http", "tcp":
		network = "tcp"
	case "https", "tls", "ssl":
		// isHTTPs = true
		network = "tcp"
	default:
		return "", nil, errors.New("listen: network '" + network + "' is unsupported")
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", nil, err
	}

	var ln net.Listener
	var lasterr error

	if port != "" && port != "0" {
		ln, err = net.Listen(network, address)
		if err == nil {
			return address, ln, nil
		}
		lasterr = err
	}

	for i := portStart; i <= portEnd; i++ {
		listenAt := net.JoinHostPort(host, strconv.Itoa(i))
		ln, err = net.Listen(network, listenAt)
		if err == nil {
			return listenAt, ln, nil
		}
		if !IsSocketBindError(err) {
			return "", nil, err
		}
		lasterr = err
	}
	if lasterr != nil {
		return "", nil, lasterr
	}
	return "", nil, errors.New("bind address fail")
}

func ParseTlsVersion(s string) (uint16, error) {
	switch s {
	case "tls10":
		return tls.VersionTLS10, nil
	case "tls11":
		return tls.VersionTLS11, nil
	case "tls12":
		return tls.VersionTLS12, nil
	case "tls13":
		return tls.VersionTLS13, nil
	default:
		return 0, errors.New("tls version '" + s + "' is invalid")
	}
}
