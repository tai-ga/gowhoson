package whoson

import (
	"expvar"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/client9/reopen"
	katsubushi "github.com/kayac/go-katsubushi"
	"go.uber.org/zap"
)

// ProtocolType is whoson protocol types.
type ProtocolType int

// MethodType is whoson protocol methods.
type MethodType int

// ResultType is whoson protocol results.
type ResultType int

// ClientConfig hold information for client configration.
type ClientConfig struct {
	Mode   string
	Server string
}

// ServerCtlConfig hold information for serverctl configration.
type ServerCtlConfig struct {
	Server     string
	JSON       bool
	EditConfig bool
}

// ServerConfig hold information for server configration.
type ServerConfig struct {
	TCP         string
	UDP         string
	Log         string
	Loglevel    string
	ServerID    int
	Expvar      bool
	ControlPort string
	SyncRemote  string
	SaveFile    string
}

const (
	/*
		" 1<<10"    1024
		" 2<<10"    2048
		" 4<<10"    4096
		" 8<<10"    8192
		"16<<10"   16384
		"32<<10"   32768
		"64<<10"   65536
		" 1<<20" 1048576 1M
	*/
	maxQueues   = 8 << 10
	udpByteSize = 1472
	charCRLF    = "\r\n"
	// SessionTimeOut is tcp session timeout limit.
	SessionTimeOut = 10 * time.Second
	// StoreDataExpire is stored data expire limit.
	StoreDataExpire = 30 * time.Minute
	// ExpireCheckInterval is expire check interval for stored data.
	ExpireCheckInterval = 5 * time.Minute

	pUnkownProtocol ProtocolType = iota
	pTCP
	pUDP

	mUnkownMethod MethodType = iota
	mLogin
	mLogout
	mQuery
	mQuit

	rPositive ResultType = iota
	rNegative
	rBadRequest
)

var (
	// MainStore holds main store.
	MainStore Store
	// Logger halds logging.
	Logger *zap.Logger
	// LogWriter is IO Writer.
	LogWriter reopen.Writer
	// IDGenerator halds id generator.
	IDGenerator *katsubushi.Generator

	// start time
	startTime = time.Now().UTC()

	// ExpvarMap halds expvar map.
	ExpvarMap = expvar.NewMap("gowhoson")

	// Raw stat collectors
	expConnectsTCPTotal   = new(expvar.Int)
	expConnectsUDPTotal   = new(expvar.Int)
	expConnectsTCPCurrent = new(expvar.Int)
	expConnectsUDPCurrent = new(expvar.Int)
	expCommandLoginTotal  = new(expvar.Int)
	expCommandLogoutTotal = new(expvar.Int)
	expCommandQueryTotal  = new(expvar.Int)
	expCommandQuitTotal   = new(expvar.Int)
	expErrorsTotal        = new(expvar.Int)

	method = map[MethodType]string{
		mUnkownMethod: "NONE",
		mLogin:        "LOGIN",
		mLogout:       "LOGOUT",
		mQuery:        "QUERY",
		mQuit:         "QUIT",
	}

	result = map[ResultType]string{
		rPositive:   "+",
		rNegative:   "-",
		rBadRequest: "*",
	}

	methodFromString = map[string]MethodType{}
)

func init() {
	for i, v := range method {
		methodFromString[v] = i
	}

	ExpvarMap.Set("ConnectsTCPTotal", expConnectsTCPTotal)
	ExpvarMap.Set("ConnectsUDPTotal", expConnectsUDPTotal)
	ExpvarMap.Set("ConnectsTCPCurrent", expConnectsTCPCurrent)
	ExpvarMap.Set("ConnectsUDPCurrent", expConnectsUDPCurrent)
	ExpvarMap.Set("CommandLoginTotal", expCommandLoginTotal)
	ExpvarMap.Set("CommandLogoutTotal", expCommandLogoutTotal)
	ExpvarMap.Set("CommandQueryTotal", expCommandQueryTotal)
	ExpvarMap.Set("CommandQuitTotal", expCommandQuitTotal)
	ExpvarMap.Set("ErrorsTotal", expErrorsTotal)
	ExpvarMap.Set("Goroutines", expvar.Func(func() interface{} { return runtime.NumGoroutine() }))
	ExpvarMap.Set("NumCPU", expvar.Func(func() interface{} { return runtime.NumCPU() }))
	ExpvarMap.Set("OSThreads", expvar.Func(func() interface{} { return pprof.Lookup("threadcreate").Count() }))
	ExpvarMap.Set("UpTime", expvar.Func(func() interface{} { return int64(time.Since(startTime)) }))
	ExpvarMap.Set("StoreCount", expvar.Func(func() interface{} { return int64(MainStore.Count()) }))
}
