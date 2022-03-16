package whoson

import (
	"fmt"
	"net"
	"net/textproto"
	"strings"

	"github.com/pkg/errors"
)

// Response hold information for response values.
type Response struct {
	result ResultType
	Msg    string
}

func (r *Response) String() string {
	return fmt.Sprintf("%s%s", result[r.result], r.Msg)
}

// Parse set values for Response.
func (r *Response) Parse(req string) error {
	switch string(req[0]) {
	case result[rPositive]:
		r.result = rPositive
	case result[rNegative]:
		r.result = rNegative
	case result[rBadRequest]:
		r.result = rBadRequest
	default:
		return errors.New("CResponse parse error")
	}
	r.Msg = string(req[1:])
	return nil
}

// Client hold information for whoson API client.
type Client struct {
	tp         *textproto.Conn
	conn       net.Conn
	serverName string
}

// Dial creates a new client connection.
func Dial(proto string, addr string) (*Client, error) {
	proto = strings.ToLower(proto)
	if proto != "tcp" && proto != "udp" {
		return nil, errors.New("Unknown protocol error")
	}
	conn, err := net.Dial(proto, addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host)
}

// NewClient return new Client struct pointer and error.
func NewClient(conn net.Conn, host string) (*Client, error) {
	tp := textproto.NewConn(conn)
	c := &Client{tp: tp, conn: conn, serverName: host}
	return c, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.tp.Close()
}

// Login access to LOGIN API.
func (c *Client) Login(ip string, args string) (*Response, error) {
	resp, err := c.doAPI("LOGIN %s %s", ip, args)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Logout access to LOGOUT API.
func (c *Client) Logout(ip string) (*Response, error) {
	resp, err := c.doAPI("LOGOUT %s", ip)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Query access to QUERY API.
func (c *Client) Query(ip string) (*Response, error) {
	resp, err := c.doAPI("QUERY %s", ip)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Quit access to QUIT API.
func (c *Client) Quit() (*Response, error) {
	resp, err := c.doAPI("QUIT")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) doAPI(format string, args ...interface{}) (*Response, error) {
	id, err := c.tp.Cmd(fmt.Sprintf("%s%s", format, charCRLF), args...)
	if err != nil {
		return nil, err
	}
	c.tp.StartResponse(id)
	defer c.tp.EndResponse(id)

	l1, err := c.tp.ReadLine()
	if err != nil {
		return nil, err
	}

	l2, err := c.tp.ReadLine()
	if err != nil {
		return nil, err
	}

	if l1 != "" && l2 == "" {
		r, err := c.newResponse(l1)

		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return nil, errors.New("Response parse error")
}

func (c *Client) newResponse(req string) (*Response, error) {
	r := &Response{}
	err := r.Parse(req)
	if err != nil {
		return nil, err
	}
	return r, err
}
