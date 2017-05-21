package whoson

import (
	"net"
	"reflect"
	"testing"
	"time"
)

var store = NewMemStore()

func newStoreData(d string) *StoreData {
	return &StoreData{
		Expire: time.Now().Add(StoreDataExpire),
		Data:   d,
	}
}

func TestNewMemStore(t *testing.T) {
	actual := reflect.TypeOf(store).String()
	expected := "whoson.MemStore"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestNewMainStore(t *testing.T) {
	NewMainStore()
	actual := reflect.TypeOf(MainStore).String()
	expected := "whoson.MemStore"
	if actual != expected {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestMemStore_SetGetDel(t *testing.T) {
	var tests = []struct {
		key      string
		value    string
		expected string
	}{
		{"key1", "value1", "value1"},
	}
	for _, tt := range tests {
		sd := newStoreData(tt.value)
		store.Set(tt.key, sd)
		actual_get, err := store.Get(tt.key)
		if err != nil {
			t.Fatalf("Error %v", err)
		}
		if tt.expected != actual_get.Data {
			t.Fatalf("expected %v, actual %v", tt.expected, actual_get)
		}
		actual_del := store.Del(tt.key)
		if err != nil {
			t.Fatalf("Error %v", err)
		}
		if actual_del != true {
			t.Fatalf("expected %v, actual %v", tt.expected, actual_del)
		}
	}
}

func TestStoreData_UpdateExpire(t *testing.T) {
	sd := newStoreData("test")
	t1 := sd.Expire
	sd.UpdateExpire()
	if !sd.Expire.After(t1) {
		t.Fatalf("t1 %v, UpdateExpire at %v", t1, sd.Expire)
	}
}

func TestStoreData_Key(t *testing.T) {
	expected := "10.0.0.1"
	sd := newStoreData("test")
	sd.IP = net.ParseIP(expected)
	actual := sd.Key()
	if expected != actual {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
