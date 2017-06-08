package whoson

import (
	"github.com/pkg/errors"
)

// ListenAndServe simple start whoson server TCP or UDP.
func ListenAndServe(proto string, addr string) error {
	switch proto {
	case "tcp":
		server := NewTCPServer()
		server.Addr = addr
		return server.ListenAndServe()
	case "udp":
		server := NewUDPServer()
		server.Addr = addr
		return server.ListenAndServe()
	default:
		return errors.New("Error ListenAndServe fail")
	}
}
