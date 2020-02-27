package whoson

import (
	"reflect"
	"testing"
)

// TestNewUDPServer is test code for NewUDPServer().
func TestNewUDPServer(t *testing.T) {
	s := NewUDPServer()
	actual := reflect.TypeOf(s).String()
	expected := "*whoson.UDPServer"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
