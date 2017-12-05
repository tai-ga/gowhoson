package whoson

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"google.golang.org/grpc"

	"github.com/olekukonko/tablewriter"
)

// ServerCtl hold information for server control.
type ServerCtl struct {
	server   string
	dumpResp *WSDumpResponse
	out      io.Writer
}

// NewServerCtl return new ServerCtl struct pointer.
func NewServerCtl(server string) *ServerCtl {
	return &ServerCtl{
		server: server,
		out:    os.Stdout,
	}
}

// Dump Set grpc repository to sc.dumpResp
func (sc *ServerCtl) Dump() error {
	l, err := grpc.Dial(sc.server,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Duration(5*time.Second)))
	if err != nil {
		return err
	}
	defer l.Close()

	client := NewSyncClient(l)

	req := &WSDumpRequest{}
	r, err := client.Dump(context.Background(), req)
	if err != nil {
		return err
	}
	sc.dumpResp = r
	return nil
}

// SetWriter Set io.Writer to sc.out
func (sc *ServerCtl) SetWriter(o io.Writer) {
	sc.out = o
}

// WriteJSON Output json with io.Writer
func (sc *ServerCtl) WriteJSON() error {
	var buf bytes.Buffer
	err := json.Indent(&buf, sc.dumpResp.Json, "", "  ")
	if err != nil {
		return err
	}
	if buf.String() != "null" {
		fmt.Fprint(sc.out, buf.String())
	}
	return nil
}

// WriteTable Output Table with io.Writer
func (sc *ServerCtl) WriteTable() error {
	t := tablewriter.NewWriter(sc.out)
	t.SetHeader([]string{"Expire", "IP", "Data"})
	t.SetAutoFormatHeaders(false)
	t.SetBorder(false)

	var sd []*StoreData
	err := json.Unmarshal(sc.dumpResp.Json, &sd)
	if err != nil {
		return err
	}

	sort.Slice(sd, func(i, j int) bool {
		return (sd[i].Expire.UnixNano() < sd[j].Expire.UnixNano()) || (sd[i].IP.String() < sd[j].IP.String())
	})

	if len(sd) > 0 {
		for _, v := range sd {
			t.Append([]string{v.Expire.Format("2006-01-02 15:04:05"), v.IP.String(), v.Data})
		}
		t.Render()
	}
	return nil
}
