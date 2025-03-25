package whoson

import (
	"fmt"
	"net"
	"net/textproto"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Session hold information for whoson session.
type Session struct {
	protocol ProtocolType
	id       uint64

	udpconn    *net.UDPConn
	remoteAddr *net.UDPAddr
	b          *Buffer

	tcpserver *TCPServer
	conn      net.Conn
	tp        *textproto.Conn
	tpid      uint

	cmdMethod MethodType
	cmdIP     net.IP
	cmdArgs   string
}

// NewSessionUDP return new Session struct pointer for UDP.
func NewSessionUDP(c *net.UDPConn, r *net.UDPAddr, b *Buffer) (*Session, error) {
	id, err := IDGenerator.NextID()
	if err != nil {
		return nil, err
	}

	return &Session{
		protocol:   pUDP,
		id:         id,
		udpconn:    c,
		remoteAddr: r,
		b:          b,
	}, nil
}

// NewSessionTCP return new Session struct pointer for TCP.
func NewSessionTCP(s *TCPServer, c net.Conn) (*Session, error) {
	id, err := IDGenerator.NextID()
	if err != nil {
		return nil, err
	}

	return &Session{
		protocol:  pTCP,
		id:        id,
		tcpserver: s,
		conn:      c,
		tp:        textproto.NewConn(c),
	}, nil
}

func (ses *Session) setTpID() {
	ses.tpid = ses.tp.Next()
}

func (ses *Session) methodType(m string) MethodType {
	if v, ok := methodFromString[m]; ok {
		return v
	}
	return mUnkownMethod
}

func (ses *Session) parseCmd(line string) error {
	var cmd []string

	if strings.TrimSpace(line) == "" {
		return errors.New("command parse error")
	}

	w := strings.Split(line, " ")
	for i := range w {
		if w[i] != "" {
			cmd = append(cmd, strings.TrimSpace(w[i]))
		}
	}

	ses.cmdMethod = ses.methodType(strings.ToUpper(cmd[0]))
	switch ses.cmdMethod {
	case mLogin, mLogout, mQuery:
		if ses.cmdIP = net.ParseIP(cmd[1]); ses.cmdIP == nil {
			return errors.New("command parse error")
		}
		ses.cmdArgs = strings.Join(cmd[2:], " ")
	case mQuit:
		ses.cmdArgs = strings.Join(cmd[1:], " ")
	default:
		return errors.New("command not found")
	}
	return nil
}

func (ses *Session) readLine() (string, error) {
	ses.tp.StartRequest(ses.tpid)
	l1, err := ses.tp.ReadLine()
	if err != nil {
		return "", err
	}

	l2, err := ses.tp.ReadLine()
	if err != nil {
		return "", err
	}
	ses.tp.EndRequest(ses.tpid)

	if l1 != "" && l2 == "" {
		return l1, nil
	}
	return "", errors.New("session read error")
}

func (ses *Session) sendLine(str string) error {
	var err error
	if ses.protocol == pTCP {
		ses.tp.StartResponse(ses.tpid)
		err = ses.tp.PrintfLine("%s%s", str, charCRLF)
		ses.tp.EndResponse(ses.tpid)
	} else {
		b := []byte(str + charCRLF + charCRLF)
		_, err = ses.udpconn.WriteToUDP(b, ses.remoteAddr)
	}
	return err
}

func (ses *Session) sendResponsePositive(str string) error {
	return ses.sendLine(fmt.Sprintf("%s%s", result[rPositive], str))
}

func (ses *Session) sendResponseNegative(str string) error {
	return ses.sendLine(fmt.Sprintf("%s%s", result[rNegative], str))
}

func (ses *Session) sendResponseBadRequest(str string) error {
	return ses.sendLine(fmt.Sprintf("%s%s", result[rBadRequest], str))
}

func (ses *Session) resetCmd() {
	ses.cmdMethod = mUnkownMethod
	ses.cmdIP = nil
	ses.cmdArgs = ""
}

func (ses *Session) startHandler() bool {
	defer ses.resetCmd()
	var err error

	if ses.protocol == pTCP {
		ses.setTpID()
		line, err := ses.readLine()
		if !ses.tcpErrorHandling(err) {
			return false
		}
		err = ses.parseCmd(line)
		if err != nil {
			Log("debug", "StartHandler", ses, err)
			ses.sendResponseBadRequest(err.Error())
			return true
		}
	} else {
		err = ses.parseCmd(string(ses.b.buf[:ses.b.count]))
		if err != nil {
			Log("debug", "StartHandler", ses, err)
			ses.sendResponseBadRequest(err.Error())
			return true
		}
	}

	switch ses.cmdMethod {
	case mLogin:
		expCommandLoginTotal.Add(1)
		ses.methodLogin()
		Log("debug", "SessionHandler", ses, err)
	case mLogout:
		expCommandLogoutTotal.Add(1)
		ses.methodLogout()
		Log("debug", "SessionHandler", ses, err)
	case mQuery:
		expCommandQueryTotal.Add(1)
		ses.methodQuery()
		Log("debug", "SessionHandler", ses, err)
	case mQuit:
		expCommandQuitTotal.Add(1)
		ses.methodQuit()
		Log("debug", "SessionHandler", ses, err)
		return false
	default:
		err := errors.New("handler error")
		expErrorsTotal.Add(1)
		Log("error", "StartHandler:Error", ses, err)
		ses.sendResponseBadRequest(err.Error())
	}
	return true
}

func (ses *Session) tcpErrorHandling(err error) bool {
	if err != nil {
		if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
			Log("debug", "StartHandler:Timeout", ses, err)
			return false
		} else if ok {
			if opError2, ok2 := opError.Err.(*os.SyscallError); ok2 && opError2.Err == syscall.ECONNRESET {
				Log("debug", "StartHandler:ResetByPeer", ses, err)
				return false
			}
		} else if err.Error() == "EOF" {
			Log("debug", "StartHandler:EOF", ses, err)
			return false
		}
		expErrorsTotal.Add(1)
		Log("error", "StartHandler:Error", ses, err)
		ses.sendResponseBadRequest(err.Error())
		return true
	}
	return true
}

func (ses *Session) methodLogin() {
	sd := &StoreData{
		Expire: time.Now().Add(StoreDataExpire),
		IP:     ses.cmdIP,
		Data:   ses.cmdArgs,
	}
	MainStore.Set(sd.Key(), sd)
	ses.sendResponsePositive("LOGIN OK")
}

func (ses *Session) methodLogout() {
	ok := MainStore.Del(ses.cmdIP.String())
	if ok {
		ses.sendResponsePositive("LOGOUT record deleted")
	} else {
		ses.sendResponsePositive("LOGOUT no such record, nothing done")
	}
}

func (ses *Session) methodQuery() {
	sd, err := MainStore.Get(ses.cmdIP.String())
	if err != nil {
		ses.sendResponseNegative("Not Logged in")
	} else {
		ses.sendResponsePositive(sd.Data)
	}
}

func (ses *Session) methodQuit() {
	ses.sendResponsePositive("QUIT OK")
	if ses.protocol == pTCP {
		ses.conn.Close()
	}
}

func (ses *Session) close() {
	if ses.protocol == pUDP {
		ses.b.Free()
	}
}
