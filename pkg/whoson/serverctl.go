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
	"google.golang.org/grpc/credentials/insecure"

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	l, err := grpc.NewClient(sc.server, opts...)
	if err != nil {
		return err
	}
	defer l.Close()

	client := NewSyncClient(l)

	req := &WSDumpRequest{}
	r, err := client.Dump(ctx, req)
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
		if sd[i].Expire.Unix() < sd[j].Expire.Unix() {
			return true
		}
		return (sd[i].Expire.Unix() == sd[j].Expire.Unix()) && (sd[i].IP.String() < sd[j].IP.String())
	})

	if len(sd) > 0 {
		for _, v := range sd {
			t.Append([]string{v.Expire.Format("2006-01-02 15:04:05"), v.IP.String(), v.Data})
		}
		t.Render()
	}
	return nil
}
