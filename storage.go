package edge

import (
	"errors"
	"fmt"
	gossh "golang.org/x/crypto/ssh"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	forwardedTCPChannelType = "forwarded-tcpip"
)

var r *rand.Rand
var rlock sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

type session struct {
	sessionId string
	alias []string
	gossh.Conn
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

var lock sync.Mutex
// sessionId -> server.Conn
var _session = make(map[string]*session)

// alias -> sessionId
var _aliasSession = make(map[string]string)

func splitHostPort(addr string) (string, uint32, error) {
	host, port, err := net.SplitHostPort(addr)
	if nil != err {
		return "", 0, err
	}
	p, _ := strconv.Atoi(port)
	return host, uint32(p), nil
}

func createConnect(alias, listenAddr, remoteAddr string) (net.Conn, error) {
	conn, err := lookupForAlias(alias)
	if nil != err {
		return nil, err
	}
	dhost, dport, err := splitHostPort(listenAddr)
	if nil != err {
		return nil, err
	}

	ohost, oport, err := splitHostPort(remoteAddr)
	if nil != err {
		return nil, err
	}

	payload := gossh.Marshal(&remoteForwardChannelData{
		DestAddr:   dhost,
		DestPort:   dport,
		OriginAddr: ohost,
		OriginPort: oport,
	})
	channel, requests, err := conn.OpenChannel(forwardedTCPChannelType, payload)
	if nil != err {
		return nil,err
	}
	go func() {
		for req := range requests {
			fmt.Println(req.Type)
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}()
	c := newConn(channel,conn.LocalAddr(),conn.RemoteAddr())
	return c,nil
}

func lookupForAlias(alias string) (gossh.Conn, error) {
	sessionId := _aliasSession[alias]
	return lookup(sessionId)
}

func lookup(sessionId string) (gossh.Conn, error) {
	if sess, found := _session[sessionId]; found {
		return sess.Conn, nil
	}
	return nil, errors.New("session not found")
}

func saveSession(sessId string, s gossh.Conn) {
	lock.Lock()
	defer lock.Unlock()

	alias := randStr(6)
	as := append(make([]string,0),alias)
	sess := &session{
		sessionId: sessId,
		Conn: s,
		alias: as,
	}
	_session[sessId] = sess
	_aliasSession[alias] = sessId
}

func getSessionAlias(sessionId string) []string {
	if sess, found := _session[sessionId]; found {
		return sess.alias
	}
	return nil
}

func deleteSession(sessionId string) {
	lock.Lock()
	defer lock.Unlock()

	if sess, found := _session[sessionId]; found {
		for _, alias := range sess.alias {
			delete(_aliasSession,alias)
		}
		if nil != sess.Conn {
			sess.Conn.Close()
		}
		delete(_session,sessionId)
	}
}

func randStr(len int) string {
	rlock.Lock()
	defer rlock.Unlock()
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		intn := r.Intn(3)
		var b int
		switch intn {
		case 0:
			b = r.Intn(26) + 65
			break
		case 1:
			b = r.Intn(26) + 97
			break
		case 2:
			b = r.Intn(10) + 48
			break
		}
		bytes[i] = byte(b)
	}
	return string(bytes)
}
