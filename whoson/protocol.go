package whoson

import (
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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
	cmdIp     net.IP
	cmdArgs   string
}

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

func (ses *Session) setTpId() {
	ses.tpid = ses.tp.Next()
}

func (ses *Session) methodType(m string) MethodType {
	if v, ok := methodFromString[m]; ok {
		return v
	} else {
		return mUnkownMethod
	}
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
		if ses.cmdIp = net.ParseIP(cmd[1]); ses.cmdIp == nil {
			return errors.New("command parse error")
		}
		ses.cmdArgs = strings.Join(cmd[2:], " ")
	case mQuit:
		//pp.Println("Quit")
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
	} else {
		return "", errors.New("session read error")
	}
}

func (ses *Session) sendLine(str string) error {
	var err error
	if ses.protocol == pTCP {
		ses.tp.StartResponse(ses.tpid)
		err = ses.tp.PrintfLine(str + CRLF)
		ses.tp.EndResponse(ses.tpid)
	} else {
		b := []byte(str + CRLF + CRLF)
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
	ses.cmdIp = nil
	ses.cmdArgs = ""
}

func (ses *Session) startHandler() bool {
	defer ses.resetCmd()
	var err error

	if ses.protocol == pTCP {
		ses.setTpId()

		line, err := ses.readLine()
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				return false
			}
			if err.Error() == "EOF" {
				return false
			} else {
				ses.sendResponseBadRequest(err.Error())
				return true
			}
		}
		err = ses.parseCmd(line)
		if err != nil {
			ses.sendResponseBadRequest(err.Error())
			return true
		}
	} else {
		err = ses.parseCmd(string(ses.b.buf[:ses.b.count]))
		if err != nil {
			ses.sendResponseBadRequest(err.Error())
			return true
		}
	}

	switch ses.cmdMethod {
	case mLogin:
		ses.methodLogin()
	case mLogout:
		ses.methodLogout()
	case mQuery:
		ses.methodQuery()
	case mQuit:
		ses.methodQuit()
	default:
		err := errors.New("handler error")
		ses.sendResponseBadRequest(err.Error())
	}
	return true
}

func (ses *Session) methodLogin() {
	sd := &StoreData{
		Expire: time.Now().Add(StoreDataExpire),
		IP:     ses.cmdIp,
		Data:   ses.cmdArgs,
	}
	MainStore.Set(sd.Key(), sd)
	ses.sendResponsePositive("LOGIN OK")
}

func (ses *Session) methodLogout() {
	ok := MainStore.Del(ses.cmdIp.String())
	if ok {
		ses.sendResponsePositive("LOGOUT record deleted")
	} else {
		ses.sendResponsePositive("LOGOUT no such record, nothing done")
	}
}

func (ses *Session) methodQuery() {
	sd, err := MainStore.Get(ses.cmdIp.String())
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
