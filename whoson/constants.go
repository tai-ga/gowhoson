package whoson

import (
	"expvar"
	"runtime"
	"runtime/pprof"
	"time"

	katsubushi "github.com/kayac/go-katsubushi"
	"go.uber.org/zap"
)

type ProtocolType int
type MethodType int
type ResultType int

type ClientConfig struct {
	Mode   string
	Server string
}

type ServerConfig struct {
	TCP string
	UDP string
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
	maxQueues       = 8 << 10
	udpByteSize     = 1472
	CRLF            = "\r\n"
	SessionTimeOut  = 10 * time.Second
	StoreDataExpire = 30 * time.Minute

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
	MainStore   Store = nil
	Logger      *zap.Logger
	IDGenerator *katsubushi.Generator

	// start time
	startTime = time.Now().UTC()

	// expvar Map
	ExpvarMap = expvar.NewMap("gowhoson")

	// Raw stat collectors
	expConnectsTcpTotal   = new(expvar.Int)
	expConnectsUdpTotal   = new(expvar.Int)
	expConnectsTcpCurrent = new(expvar.Int)
	expConnectsUdpCurrent = new(expvar.Int)
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

	ExpvarMap.Set("ConnectsTcpTotal", expConnectsTcpTotal)
	ExpvarMap.Set("ConnectsUdpTotal", expConnectsUdpTotal)
	ExpvarMap.Set("ConnectsTcpCurrent", expConnectsTcpCurrent)
	ExpvarMap.Set("ConnectsUdpCurrent", expConnectsUdpCurrent)
	ExpvarMap.Set("CommandLoginTotal", expCommandLoginTotal)
	ExpvarMap.Set("CommandLogoutTotal", expCommandLogoutTotal)
	ExpvarMap.Set("CommandQueryTotal", expCommandQueryTotal)
	ExpvarMap.Set("CommandQuitTotal", expCommandQuitTotal)
	ExpvarMap.Set("ErrorsTotal", expErrorsTotal)
	ExpvarMap.Set("Goroutines", expvar.Func(func() interface{} { return runtime.NumGoroutine() }))
	ExpvarMap.Set("NumCPU", expvar.Func(func() interface{} { return runtime.NumCPU() }))
	ExpvarMap.Set("OSThreads", expvar.Func(func() interface{} { return pprof.Lookup("threadcreate").Count() }))
	ExpvarMap.Set("UpTime", expvar.Func(func() interface{} { return int64(time.Since(startTime)) }))
}
