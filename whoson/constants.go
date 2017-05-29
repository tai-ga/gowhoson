package whoson

import "time"

type ProtocolType int
type MethodType int
type ResultType int

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
	MainStore Store = nil

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
}
