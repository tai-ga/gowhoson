package whoson

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
)

// Store is hold Store API.
type Store interface {
	Set(k string, w *StoreData)
	Get(k string) (*StoreData, error)
	Del(k string) bool
	Items() map[string]interface{}
	Count() int
}

// MemStore hold information for cmap.
type MemStore struct {
	cmap cmap.ConcurrentMap
	Store
}

// NewMemStore return new MemStore.
func NewMemStore() Store {
	return MemStore{
		cmap: cmap.New(),
	}
}

// NewMainStore set MemStore to MainStore.
func NewMainStore() {
	if MainStore == nil {
		MainStore = NewMemStore()
	}
}

// Set data to cmap store.
func (ms MemStore) Set(k string, w *StoreData) {
	ms.cmap.Set(k, w)
}

// Get data from cmap store.
func (ms MemStore) Get(k string) (*StoreData, error) {
	if v, ok := ms.cmap.Get(k); ok {
		if w, ok := v.(*StoreData); ok {
			if w.Expire.After(time.Now()) {
				return w, nil
			}
			ms.Del(k)
			return nil, errors.New("data not found")
		}
		return nil, errors.New("type assertion error")
	}
	return nil, errors.New("data not found")
}

// Del delete data from cmap store.
func (ms MemStore) Del(k string) bool {
	if ms.cmap.Has(k) {
		ms.cmap.Remove(k)
		return true
	}
	return false
}

// Items return all data from cmap store.
func (ms MemStore) Items() map[string]interface{} {
	return ms.cmap.Items()
}

// Count return all data size.
func (ms MemStore) Count() int {
	return ms.cmap.Count()
}

// StoreData hold information for whoson data.
type StoreData struct {
	Expire time.Time
	IP     net.IP
	Data   string
}

// UpdateExpire Update stored data of expire time.
func (sd *StoreData) UpdateExpire() {
	sd.Expire = time.Now().Add(StoreDataExpire)
}

// Key return key string.
func (sd *StoreData) Key() string {
	return sd.IP.String()
}

func deleteExpireData(store Store) {
	for k, v := range store.Items() {
		if w, ok := v.(*StoreData); ok {
			if w.Expire.Before(time.Now()) {
				msg := fmt.Sprintf("ExpireData:%s", k)
				Log("info", msg, nil, nil)
				store.Del(k)
			}
		}
	}
}

// RunExpireChecker Check expire for all cmap store data.
func RunExpireChecker(ctx context.Context) {
	t := time.NewTicker(ExpireCheckInterval)
	Log("info", "runExpireCheckerStart", nil, nil)
	for {
		select {
		case <-ctx.Done():
			Log("info", "runExpireCheckerStop", nil, nil)
			return
		case <-t.C:
			if MainStore != nil {
				deleteExpireData(MainStore)
			}
		}
	}
}
