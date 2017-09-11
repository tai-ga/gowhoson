package whoson

import (
	"net"
	"time"

	"golang.org/x/net/context"
)

// Sync hold information for synchronization.
type Sync struct{}

// Set sync to repliction servers
func (s *Sync) Set(c context.Context, wreq *WSRequest) (*WSResponse, error) {
	ip := net.ParseIP(wreq.IP)
	req := &StoreData{
		Expire: time.Unix(wreq.Expire, 0),
		IP:     ip,
		Data:   wreq.Data,
	}
	MainStore.SyncSet(ip.String(), req)
	return &WSResponse{Msg: "OK", Rcode: 1}, nil
}

// Del delete to repliction servers
func (s *Sync) Del(c context.Context, wreq *WSRequest) (*WSResponse, error) {
	ip := net.ParseIP(wreq.IP)
	if MainStore.SyncDel(ip.String()) {
		return &WSResponse{Msg: "OK", Rcode: 1}, nil
	}
	return &WSResponse{Msg: "NG", Rcode: 2}, nil
}
