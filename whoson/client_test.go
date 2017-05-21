package whoson

import (
	"net"
	"sync"
	"testing"
)

type TestEnv struct {
	udpconn   *net.UDPConn
	udpclient *Client

	tcpconn   *net.TCPListener
	tcpclient *Client
	wg        *sync.WaitGroup
}

func (te *TestEnv) close() {
	te.udpconn.Close()
	te.tcpconn.Close()
	te.wg.Wait()
}

var te TestEnv

func startUDPServer(t *testing.T) {
	var err error
	if te.wg == nil {
		te.wg = new(sync.WaitGroup)
	}

	if te.udpconn == nil {
		addrudp := net.UDPAddr{
			Port: 9876,
			IP:   net.ParseIP("localhost"),
		}
		te.udpconn, err = net.ListenUDP("udp", &addrudp)
		if err != nil {
			t.Fatalf("Error %v", err)
			return
		}

		te.wg.Add(1)
		go func() {
			defer te.wg.Done()
			ServeUDP(te.udpconn)
		}()
	}

	if te.tcpconn == nil {
		addrtcp := net.TCPAddr{
			Port: 9876,
			IP:   net.ParseIP("localhost"),
		}
		te.tcpconn, err = net.ListenTCP("tcp", &addrtcp)
		if err != nil {
			t.Fatalf("Error %v", err)
			return
		}
		te.wg.Add(1)
		go func() {
			defer te.wg.Done()
			ServeTCP(te.tcpconn)
		}()
	}

	if te.udpclient == nil {
		te.udpclient, err = Dial("udp", "localhost:9876")
		if err != nil {
			t.Fatalf("Error %v", err)
			return
		}
	}

	if te.tcpclient == nil {
		te.tcpclient, err = Dial("tcp", "localhost:9876")
		if err != nil {
			t.Fatalf("Error %v", err)
			return
		}
	}
	return
}

func TestClient_Commands(t *testing.T) {
	startUDPServer(t)

	var r *Response
	var err error
	var tests = []struct {
		protocol string
		method   string
		args1    string
		args2    string
		expected string
	}{
		{"udp", "login", "1.1.1", "TESTSTRING", "*command parse error"},
		{"udp", "login", "1.1.1.1", "TESTSTRING", "+LOGIN OK"},
		{"udp", "query", "1.1.1.1", "", "+TESTSTRING"},
		{"udp", "query", "1.1.1.2", "", "-Not Logged in"},
		{"udp", "logout", "1.1.1.1", "", "+LOGOUT record deleted"},
		{"udp", "logout", "1.1.1.1", "", "+LOGOUT no such record, nothing done"},
		{"tcp", "login", "1.1.1", "TESTSTRING", "*command parse error"},
		{"tcp", "login", "1.1.1.1", "TESTSTRING", "+LOGIN OK"},
		{"tcp", "query", "1.1.1.1", "", "+TESTSTRING"},
		{"tcp", "query", "1.1.1.2", "", "-Not Logged in"},
		{"tcp", "logout", "1.1.1.1", "", "+LOGOUT record deleted"},
		{"tcp", "logout", "1.1.1.1", "", "+LOGOUT no such record, nothing done"},
		{"udp", "login", "2.2.2.2", "TESTSTRING2", "+LOGIN OK"},
		{"tcp", "query", "2.2.2.2", "", "+TESTSTRING2"},
		{"udp", "logout", "2.2.2.2", "", "+LOGOUT record deleted"},
		{"udp", "quit", "", "", "+QUIT OK"},
		{"tcp", "quit", "", "", "+QUIT OK"},
	}
	for _, tt := range tests {
		switch tt.method {
		case "login":
			if tt.protocol == "udp" {
				r, err = te.udpclient.Login(tt.args1, tt.args2)
			} else {
				r, err = te.tcpclient.Login(tt.args1, tt.args2)
			}
		case "logout":
			if tt.protocol == "udp" {
				r, err = te.udpclient.Logout(tt.args1)
			} else {
				r, err = te.tcpclient.Logout(tt.args1)
			}
		case "query":
			if tt.protocol == "udp" {
				r, err = te.udpclient.Query(tt.args1)
			} else {
				r, err = te.tcpclient.Query(tt.args1)
			}
		case "quit":
			if tt.protocol == "udp" {
				r, err = te.udpclient.Quit()
			} else {
				r, err = te.tcpclient.Quit()
			}
		default:
			t.Fatalf("Error command not found")
		}
		if err != nil {
			t.Fatalf("Error %v", err)
		}
		actual := r.String()
		if tt.expected != actual {
			t.Fatalf("expected: %v, actual: %v, protocol: %v, method: %v", tt.expected, actual, tt.protocol, tt.method)
		}
	}
	te.close()
}
