package gowhoson

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/tai-ga/gowhoson/pkg/whoson"
	"github.com/urfave/cli/v2"
)

type TestEnv struct {
	tcpconn *net.TCPListener
	wg      *sync.WaitGroup
}

func (te *TestEnv) close() {
	te.tcpconn.Close()
	te.wg.Wait()
}

var te TestEnv

func startTCPServer(t *testing.T) {
	var err error
	if te.wg == nil {
		te.wg = new(sync.WaitGroup)
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
			whoson.ServeTCP(te.tcpconn)
		}()
	}
}

func testWithServer(t *testing.T, testFuncs ...func(*cli.App)) string {
	AppVersions = NewVersions("", "")
	var buf bytes.Buffer
	startTCPServer(t)

	app := makeApp()
	app.Writer = &buf
	app.Metadata = map[string]interface{}{
		"config": &whoson.ClientConfig{
			Mode:   "tcp",
			Server: "localhost:9876",
		},
	}

	for _, f := range testFuncs {
		f(app)
	}
	te.close()
	return buf.String()
}

func TestCmd(t *testing.T) {
	out := testWithServer(
		t,
		func(app *cli.App) {
			app.Run([]string{"gowhoson", "client", "login", "1.1.1.1", "TESTSTRING"})
		},
		func(app *cli.App) {
			app.Run([]string{"gowhoson", "client", "query", "1.1.1.1"})
		},
		func(app *cli.App) {
			app.Run([]string{"gowhoson", "client", "logout", "1.1.1.1"})
		},
	)
	for _, s := range []string{"+LOGIN OK", "+TESTSTRING", "+LOGOUT record deleted"} {
		if !strings.Contains(out, s) {
			t.Fatalf("%q should be contained in output of command: %v", s, out)
		}
	}
}
