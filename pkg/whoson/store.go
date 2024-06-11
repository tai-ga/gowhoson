package whoson

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var syncChan chan *WSRequest

// Store is hold Store API.
type Store interface {
	Set(k string, w *StoreData)
	Get(k string) (*StoreData, error)
	Del(k string) bool
	Items() map[string]*StoreData
	ItemsJSON() ([]byte, error)
	Count() int
	SyncSet(k string, w *StoreData)
	SyncDel(k string) bool
}

var _ Store = (*MemStore)(nil)

// MemStore hold information for cmap.
type MemStore struct {
	cmap       cmap.ConcurrentMap[string, *StoreData]
	SyncRemote bool
	Store
}

// NewMemStore return new MemStore.
func NewMemStore() Store {
	return MemStore{
		cmap:       cmap.New[*StoreData](),
		SyncRemote: false,
	}
}

// NewMainStore set MemStore to MainStore.
func NewMainStore() {
	if MainStore == nil {
		MainStore = NewMemStore()
	}
}

// NewMainStoreEnableSyncRemote set MemStore to MainStore, enable sync remote.
func NewMainStoreEnableSyncRemote() {
	if MainStore == nil {
		MainStore = MemStore{
			cmap:       cmap.New[*StoreData](),
			SyncRemote: true,
		}
	}
	if syncChan == nil {
		syncChan = make(chan *WSRequest, 32)
	}
}

// Set data to cmap store.
func (ms MemStore) Set(k string, w *StoreData) {
	ms.cmap.Set(k, w)

	if ms.SyncRemote {
		r := &WSRequest{
			Expire: w.Expire.Unix(),
			IP:     w.IP.String(),
			Data:   w.Data,
			Method: "Set",
		}
		syncChan <- r
	}
}

// SyncSet data to remote host store.
func (ms MemStore) SyncSet(k string, w *StoreData) {
	ms.cmap.Set(k, w)
}

// Get data from cmap store.
func (ms MemStore) Get(k string) (*StoreData, error) {
	if item, ok := ms.cmap.Get(k); ok {
		if item.Expire.After(time.Now()) {
			return item, nil
		}
		ms.SyncDel(k)
	}
	return nil, errors.New("data not found")
}

// Del delete data from cmap store.
func (ms MemStore) Del(k string) bool {
	if ms.SyncRemote {
		r := &WSRequest{
			IP:     k,
			Method: "Del",
		}
		syncChan <- r
	}

	if ms.cmap.Has(k) {
		ms.cmap.Remove(k)
		return true
	}
	return false
}

// SyncDel data from remote host store.
func (ms MemStore) SyncDel(k string) bool {
	if ms.cmap.Has(k) {
		ms.cmap.Remove(k)
		return true
	}
	return false
}

// Items return all data from cmap store.
func (ms MemStore) Items() map[string]*StoreData {
	return ms.cmap.Items()
}

// ItemsJSON return all data of json format.
func (ms MemStore) ItemsJSON() ([]byte, error) {
	var sd []*StoreData
	for _, item := range ms.Items() {
		if item.Expire.Before(time.Now()) {
			msg := fmt.Sprintf("ExpireData:%s", item.Key())
			Log("info", msg, nil, nil)
			if MainStore != nil {
				MainStore.SyncDel(item.Key())
			}
		} else {
			sd = append(sd, item)
		}
	}
	return json.Marshal(sd)
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
	for _, item := range store.Items() {
		if item.Expire.Before(time.Now()) {
			msg := fmt.Sprintf("ExpireData:%s", item.Key())
			Log("info", msg, nil, nil)
			store.Del(item.Key())
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

// RunSyncRemote is sync data to remote grpc servers.
func RunSyncRemote(ctx context.Context, hosts []string) {
	defer close(syncChan)

	if Logger != nil {
		grpc_zap.ReplaceGrpcLogger(Logger)
	}
	Log("info", "RunSyncRemoteStart", nil, nil)
	for {
		select {
		case <-ctx.Done():
			Log("info", "RunSyncRemoteStop", nil, nil)
			return
		case req, ok := <-syncChan:
			if !ok {
				return
			}
			for _, h := range hosts {
				if h != "" {
					go execSyncRemote(req, h)
				}
			}
		}
	}
}

func execSyncRemote(req *WSRequest, remotehost string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.NewClient(remotehost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		Log("error", "execSyncRemote:Error", nil, err)
		return
	}
	defer conn.Close()

	client := NewSyncClient(conn)

	switch req.Method {
	case "Set":
		_, err = client.Set(ctx, req)
		Log("debug", "execSyncRemote:Set", nil, nil)
	case "Del":
		_, err = client.Del(ctx, req)
		Log("debug", "execSyncRemote:Del", nil, nil)
	}
	if err != nil {
		Log("error", "execSyncRemote:Error", nil, err)
	}
}
