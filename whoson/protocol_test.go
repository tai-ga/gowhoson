package whoson

import (
	"net"
	"reflect"
	"testing"
)

func TestNewSessionUDP(t *testing.T) {
	bp := NewBufferPool()
	s := NewSessionUDP(new(net.UDPConn), new(net.UDPAddr), bp.Get())
	actual := reflect.TypeOf(s).String()
	expected := "*whoson.Session"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestNewSessionTCP(t *testing.T) {
}
