package loong

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/mei-rune/ipfilter"
	"github.com/runner-mei/errors"
	"github.com/runner-mei/log"
)

var ErrServerInitializing = errors.New("service initializing")
var ErrServerAlreadyStart = errors.New("service already start")
var ErrServerAlreadyStop = errors.New("service already stop")

type Hook interface {
	OnStart(context.Context, *Runner) error
	OnStop(context.Context, *Runner) error
}
type hook struct {
	onStart func(context.Context, *Runner) error
	onStop  func(context.Context, *Runner) error
}

func (h hook) OnStart(ctx context.Context, r *Runner) error {
	if h.onStart == nil {
		return nil
	}

	return h.onStart(ctx, r)
}
func (h hook) OnStop(ctx context.Context, r *Runner) error {
	if h.onStop == nil {
		return nil
	}

	return h.onStop(ctx, r)
}

func MakeHook(onStart, onStop func(context.Context, *Runner) error) Hook {
	return hook{
		onStart: onStart,
		onStop:  onStop,
	}
}

type Runner struct {
	Logger          log.Logger
	IPFilterOptions ipfilter.Options
	Network         string
	ListenAt        string

	KeyFile  string
	CertFile string

	TLCP struct {
		SigCertFile string
		SigKeyFile  string
		EncCertFile string
		EncKeyFile  string
	}

	CandidatePortStart int
	CandidatePortEnd   int

	lock     sync.Mutex
	srv      *http.Server
	listener net.Listener

	hooks []Hook
}

func (r *Runner) Append(hook Hook) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.hooks = append(r.hooks, hook)
}

func (r *Runner) MustURL(address ...string) string {
	u, err := r.URL()
	if err != nil {
		panic(err)
	}
	return u
}

func (r *Runner) URL(address ...string) (string, error) {
	port, err := r.ListenPort()
	if err != nil {
		return "", err
	}

	var hostAddress = "127.0.0.1"
	if len(address) > 0 {
		if !isZeroAddress(address[0]) {
			hostAddress = address[0]
		}
	}

	network := r.Network
	switch strings.ToLower(network) {
	case "http", "tcp":
		network = "http"
	case "https", "tls", "ssl", "tlcp":
		network = "https"
	default:
		return "", errors.New("network '" + network + "' is unsupported")
	}
	return network + "://" + net.JoinHostPort(hostAddress, port), nil
}

func (r *Runner) ListenAddr() (net.Addr, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.srv == nil {
		return nil, ErrServerInitializing
	}
	return r.listener.Addr(), nil
}

func isZeroAddress(addr string) bool {
	return addr == "" || addr == "[::]" || addr == ":" || addr == ":0" || addr == "0.0.0.0:0"
}

func (r *Runner) ListenPort() (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.srv == nil {
		return "", ErrServerInitializing
	}

	// if isZeroAddress(r.ListenAt) {
	_, port, err := net.SplitHostPort(r.listener.Addr().String())
	return port, err
	// }
	// _, port, err := net.SplitHostPort(r.ListenAt)
	// return port, err
}

func (r *Runner) Run(ctx context.Context, handler http.Handler) error {
	stopped := make(chan struct{})
	err := r.start(ctx, handler, stopped)
	if err != nil {
		return err
	}

	select {
	case <-stopped:
	case <-ctx.Done():
	}

	return r.Stop(ctx)
}

func (r *Runner) Start(ctx context.Context, handler http.Handler) error {
	return r.start(ctx, handler, nil)
}

func (r *Runner) start(ctx context.Context, handler http.Handler, stopped chan struct{}) error {
	if handler == nil {
		return errors.New("handler is missing")
	}
	network := r.Network
	isHTTPs := false
	isTLCP := false

	switch strings.ToLower(network) {
	case "http", "tcp":
		network = "tcp"
	case "https", "tls", "ssl":
		isHTTPs = true
		network = "tcp"
		if r.CertFile == "" || r.KeyFile == "" {
			return errors.New("keyFile or certFile is missing")
		}
	case "tlcp":
		isTLCP = true
		isHTTPs = true
		network = "tcp"
		if r.TLCP.SigCertFile == "" || r.TLCP.SigKeyFile == "" {
			return errors.New("sig keyFile or certFile is missing")
		}
		if r.TLCP.EncCertFile == "" || r.TLCP.EncKeyFile == "" {
			return errors.New("enc keyFile or certFile is missing")
		}
	default:
		return errors.New("listen: network '" + network + "' is unsupported")
	}

	var srv *http.Server
	var listener net.Listener
	var hooks []Hook

	err := func() error {
		r.lock.Lock()
		defer r.lock.Unlock()

		if r.srv != nil {
			return ErrServerAlreadyStart
		}

		listenAt, ln, err := ListenAtDynamicPort(network, r.ListenAt, r.CandidatePortStart, r.CandidatePortEnd)
		if err != nil {
			return err
		}

		listener = ln
		srv = &http.Server{Addr: listenAt, Handler: handler}

		r.listener = listener
		r.srv = srv

		hooks = make([]Hook, len(r.hooks))
		copy(hooks, r.hooks)
		return nil
	}()
	if err != nil {
		return err
	}

	for idx := range hooks {
		err = hooks[idx].OnStart(ctx, r)
		if err != nil {
			listener.Close()

			for i := 0; i < idx; i++ {
				hooks[i].OnStop(ctx, r)
			}
			return err
		}
	}

	go func() {
		if stopped != nil {
			defer close(stopped)
		}

		r.Logger.Info("http listen at: " + r.Network + "+" + listener.Addr().String())

		tcpListener, ok := listener.(*net.TCPListener)
		if ok {
			listener = TcpKeepAliveListener{tcpListener}
		}

		if !r.IPFilterOptions.TrustProxy {
			listener = ipfilter.WrapListener(listener, r.IPFilterOptions, func(addr net.Addr) {
				r.Logger.Info("ip is blocked", log.Stringer("addr", addr))
			})
		}

		var err error
		if isTLCP {
			listener, err = r.enableTlcp(listener)
			if err != nil {
				r.Logger.Error("enable tlcp unsuccessful", log.Error(err))
				err = errors.Wrap(err, "enable tlcp unsuccessful")
			} else {
				err = srv.Serve(listener)
			}
		} else if isHTTPs {
			err = srv.ServeTLS(listener, r.CertFile, r.KeyFile)
		} else {
			err = srv.Serve(listener)
		}
		if err != nil {
			if err != http.ErrServerClosed {
				r.Logger.Error("http server start unsuccessful", log.Error(err))
			} else {
				r.Logger.Info("http server stopped")
			}
		}
	}()
	return nil
}

func (r *Runner) Stop(ctx context.Context) error {
	hooks, err := func() ([]Hook, error) {
		r.lock.Lock()
		defer r.lock.Unlock()

		if r.srv == nil {
			return nil, nil
		}

		listenAt := r.listener.Addr().String()

		err1 := r.srv.Close()
		err2 := r.listener.Close()
		if err2 != nil {
			if strings.Contains(err2.Error(), "use of closed network connection") {
				err2 = nil
			}
		}

		r.srv = nil
		r.listener = nil
		if err := errors.Join(err1, err2); err != nil {
			r.Logger.Info("http '" + r.Network + "+" + listenAt + "' is stop failure")
			return nil, err
		}

		r.Logger.Info("http '" + r.Network + "+" + listenAt + "' is stopped")

		hooks := make([]Hook, len(r.hooks))
		copy(hooks, r.hooks)
		return hooks, nil
	}()
	if err != nil {
		return err
	}

	for idx := range hooks {
		err = hooks[idx].OnStop(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}
