package whoson

import (
	"context"
	"fmt"
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
	NewMainStore()
	if s.listener == nil {
		s.listener = l
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	var err error
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
	fmt.Println("TCP Core: done")
	return err
}

func (s *TCPServer) startSession(ctx context.Context, conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	ses := NewSessionTCP(s, conn)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("TCP Worker: done")
			return
		default:
		}

		if err := conn.SetDeadline(time.Now().Add(s.timeOut)); err != nil {
			err = errors.Wrap(err, "Can't set appropriate deadline!")
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
	//log.Tracef("whoson process waited")
}
