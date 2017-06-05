package whoson

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type TCPServer struct {
	listener *net.TCPListener
	timeOut  time.Duration
	wg       *sync.WaitGroup
	Addr     string
}

func NewTCPServer() *TCPServer {
	return &TCPServer{
		timeOut: SessionTimeOut,
		wg:      &sync.WaitGroup{},
	}
}

func ServeTCP(l *net.TCPListener) error {
	s := NewTCPServer()
	s.listener = l
	return s.ServeTCP(l)
}

func (s *TCPServer) ListenAndServe() error {
	var addrudp net.TCPAddr
	addr := s.Addr
	if addr == "" {
		addrudp = net.TCPAddr{
			Port: 9876,
			IP:   net.ParseIP("0.0.0.0"),
		}
	} else {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		addrudp = net.TCPAddr{
			Port: p,
			IP:   net.ParseIP(host),
		}
	}
	l, err := net.ListenTCP("tcp", &addrudp)
	if err != nil {
		return err
	}
	return s.ServeTCP(l)
}

func (s *TCPServer) ServeTCP(l *net.TCPListener) error {
	var err error
	NewMainStore()
	NewLogger("stdout", "warn")
	err = NewIDGenerator(uint32(1))
	if err != nil {
		return errors.Wrap(err, "IDGenerator failed")
	}

	Log("info", "TCPServerStart", nil, nil)
	if s.listener == nil {
		s.listener = l
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			err = errors.New("gowhoson: Core closed")
			goto DONE
		default:
		}
		conn, err := s.listener.Accept()
		if err != nil {
			goto DONE
		}

		s.wg.Add(1)
		go s.startSession(ctx, conn)
	}
DONE:
	ctxCancel()
	s.wait()
	Log("info", "TCPServerStop", nil, nil)
	return err
}

func (s *TCPServer) startSession(ctx context.Context, conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
		expConnectsTcpCurrent.Add(-1)
	}()

	expConnectsTcpTotal.Add(1)
	expConnectsTcpCurrent.Add(1)
	ses, err := NewSessionTCP(s, conn)
	if err != nil {
		expErrorsTotal.Add(1)
		Log("error", "Session failed", ses, err)
	}
	Log("debug", "Session start", ses, nil)
	for {
		select {
		case <-ctx.Done():
			Log("info", "TCPServerWorkerStop", nil, nil)
			return
		default:
		}

		if err := conn.SetDeadline(time.Now().Add(s.timeOut)); err != nil {
			Log("error", "startSession:Error", ses, err)
			return
		}

		if !ses.startHandler() {
			return
		}
	}
}

// wait causes the caller to block until all active Whoson sessions have finished
func (s *TCPServer) wait() {
	s.wg.Wait()
}
