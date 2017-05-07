package whoson

import (
	"net"
	"time"

	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
)

type Store interface {
	Set(k string, w *StoreData)
	Get(k string) (*StoreData, error)
	Del(k string) bool
}

type MemStore struct {
	cmap cmap.ConcurrentMap
	Store
}

func NewMemStore() Store {
	return MemStore{
		cmap: cmap.New(),
	}
}

func NewMainStore() Store {
	if MainStore == nil {
		MainStore = NewMemStore()
	}
	return MainStore
}

func (ms MemStore) Set(k string, w *StoreData) {
	ms.cmap.Set(k, w)
}

func (ms MemStore) Get(k string) (*StoreData, error) {
	if v, ok := ms.cmap.Get(k); ok {
		if w, ok := v.(*StoreData); ok {
			if w.Expire.After(time.Now()) {
				return w, nil
			} else {
				ms.Del(k)
				return nil, errors.New("data not found")
			}
		}
		return nil, errors.New("type assertion error")
	}
	return nil, errors.New("data not found")
}

func (ms MemStore) Del(k string) bool {
	if ms.cmap.Has(k) {
		ms.cmap.Remove(k)
		return true
	} else {
		return false
	}
}

type StoreData struct {
	Expire time.Time
	IP     net.IP
	Data   string
}

func (sd *StoreData) UpdateExpire() {
	sd.Expire = time.Now().Add(StoreDataExpire)
}

func (sd *StoreData) Key() string {
	return sd.IP.String()
}
