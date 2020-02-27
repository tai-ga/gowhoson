package whoson

import (
	"reflect"
	"testing"
)

func TestNewTCPServer(t *testing.T) {
	s := NewTCPServer()
	actual := reflect.TypeOf(s).String()
	expected := "*whoson.TCPServer"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
