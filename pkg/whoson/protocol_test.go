package whoson

import (
	"net"
	"reflect"
	"testing"
)

func TestNewSessionUDP(t *testing.T) {
	bp := NewBufferPool()
	s, err := NewSessionUDP(new(net.UDPConn), new(net.UDPAddr), bp.Get())
	if err != nil {
		t.Fatalf("Error %v", err)
		return
	}
	actual := reflect.TypeOf(s).String()
	expected := "*whoson.Session"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestNewSessionTCP(t *testing.T) {
}
