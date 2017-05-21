package whoson

import (
	"reflect"
	"testing"
)

var bp = NewBufferPool()

func TestNewBufferPool(t *testing.T) {
	actual := reflect.TypeOf(bp).String()
	expected := "*whoson.BufferPool"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestBuffer_Free(t *testing.T) {
	b := bp.Get()
	b.Free()
}

func TestBufferPool_Get(t *testing.T) {
	actual := reflect.TypeOf(bp.Get()).String()
	expected := "*whoson.Buffer"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
