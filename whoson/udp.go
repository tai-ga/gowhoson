package whoson

import (
	"context"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// UDPServer hold information for udp server.
type UDPServer struct {
	conn    *net.UDPConn
	bp      *BufferPool
	queue   chan interface{}
	workers []*Worker
	timeOut time.Duration
	wg      *sync.WaitGroup
	Addr    string
}

// NewUDPServer return new UDPServer struct pointer.
func NewUDPServer() *UDPServer {
	return &UDPServer{
		bp:      NewBufferPool(),
		queue:   make(chan interface{}, maxQueues),
		timeOut: SessionTimeOut,
		wg:      &sync.WaitGroup{},
	}
}

func (s *UDPServer) enqueue(v interface{}) {
	s.queue <- v
}

// ServeUDP is start udp server serve.
func ServeUDP(c *net.UDPConn) error {
	s := NewUDPServer()
	s.conn = c
	return s.ServeUDP(c)
}

//ListenAndServe simple start udp server.
func (s *UDPServer) ListenAndServe() error {
	var addrudp net.UDPAddr
	addr := s.Addr
	if addr == "" {
		addrudp = net.UDPAddr{
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
		addrudp = net.UDPAddr{
			Port: p,
			IP:   net.ParseIP(host),
		}
	}
	c, err := net.ListenUDP("udp", &addrudp)
	if err != nil {
		return err
	}
	return s.ServeUDP(c)
}

// ServeUDP is start udp server serve.
func (s *UDPServer) ServeUDP(c *net.UDPConn) error {
	var err error

	NewMainStore()
	NewLogger("stdout", "warn")
	err = NewIDGenerator(uint(1))
	if err != nil {
		return errors.Wrap(err, "IDGenerator failed")
	}

	if s.conn == nil {
		s.conn = c
	}
	ctx, ctxCancel := context.WithCancel(context.Background())

	maxWorkers := runtime.NumCPU()
	s.workers = make([]*Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		w := Worker{
			s: s,
		}
		s.workers[i] = &w
	}
	for _, w := range s.workers {
		s.wg.Add(1)
		go w.Run(ctx)
	}
	err = s.startSession(ctx)
	ctxCancel()
	s.wg.Wait()
	return err
}

func (s *UDPServer) getBuffer() *Buffer {
	return s.bp.Get()
}

func (s *UDPServer) startSession(ctx context.Context) error {
	var n int
	var a *net.UDPAddr
	var err error
	for {
		select {
		case <-ctx.Done():
			err = errors.New("gowhoson: Core closed")
			goto DONE
		default:
		}

		if err := s.conn.SetDeadline(time.Now().Add(s.timeOut)); err != nil {
			err = errors.Wrap(err, "Can't set appropriate deadline!")
			return err
		}

		b := s.getBuffer()
		n, a, err = s.conn.ReadFromUDP(b.buf)
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				b.Free()
				continue
			}
			goto DONE
		}
		b.count = n
		ses, err := NewSessionUDP(s.conn, a, b)
		if err != nil {
			expErrorsTotal.Add(1)
			Log("error", "Session failed", ses, err)
		}

		s.enqueue(ses)
		Log("debug", "Session start", ses, nil)
	}
DONE:
	Log("info", "UDPServerStop", nil, nil)
	return err
}

// wait causes the caller to block until all active Whoson sessions have finished
func (s *UDPServer) wait() {
	s.wg.Wait()
}

// Worker hold information for udp server processing workers.
type Worker struct {
	s *UDPServer
}

// Run start worker processing.
func (w *Worker) Run(ctx context.Context) {
	defer w.s.wg.Done()

	for {
		select {
		case v := <-w.s.queue:
			w.work(v)
		case <-ctx.Done():
			Log("info", "UDPServerWorkerStop", nil, nil)
			return
		}
	}
}

func (w *Worker) work(v interface{}) {
	if ses, ok := v.(*Session); ok {
		defer func() {
			ses.close()
			expConnectsUDPCurrent.Add(-1)
		}()
		expConnectsUDPTotal.Add(1)
		expConnectsUDPCurrent.Add(1)
		ses.startHandler()
	} else {
		panic(v)
	}
}
