package edge

import (
	gossh "golang.org/x/crypto/ssh"
	"net"
	"time"
)

type Conn struct {
	gossh.Channel
	local  net.Addr
	remote net.Addr
}

func newConn(s gossh.Channel, local, remote net.Addr) *Conn {
	return &Conn{
		Channel: s,
		local:   local,
		remote:  remote}
}

func (c *Conn) LocalAddr() net.Addr {
	return c.local
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

