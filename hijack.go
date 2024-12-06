package echogy

import (
	"bufio"
	"bytes"
	q "github.com/youkale/echogy/pkg/queue"
	"net"
	"net/http"
	"time"
)

type Dispatch func(*http.Response, *http.Request, int64)

type hijackConn struct {
	net.Conn
	dispatch Dispatch
	q        *q.SyncQueue
}

type request struct {
	*http.Request
	startTime int64
}

func newHijackConn(conn net.Conn) *hijackConn {
	return &hijackConn{
		Conn: conn,
		q:    q.NewSyncQueue(16),
	}
}

func (h *hijackConn) AddRequest(r *http.Request) {
	h.q.Push(&request{
		Request:   r,
		startTime: time.Now().UnixMilli(),
	})
}

func (h *hijackConn) Read(b []byte) (n int, err error) {
	n, err = h.Conn.Read(b)
	if err != nil {
		return n, err
	}

	// Try to parse HTTP request from the data
	reader := bytes.NewReader(b[:n])
	if req, err := http.ReadRequest(bufio.NewReader(reader)); err == nil {
		// add req to queue
		h.q.Push(&request{
			Request:   req,
			startTime: time.Now().UnixMilli(),
		})
	}
	return n, nil
}

func (h *hijackConn) SetDispatch(d Dispatch) {
	h.dispatch = d
}

func (h *hijackConn) Write(b []byte) (n int, err error) {
	n, err = h.Conn.Write(b)
	if nil != err {
		return n, err
	}

	pop := h.q.Pop()
	if nil != pop {
		r := pop.(*request)
		// pop request from request queue and Try to parse HTTP response from the data
		if nil != h.dispatch {
			reader := bytes.NewReader(b)
			if resp, err := http.ReadResponse(bufio.NewReader(reader), r.Request); nil == err {
				useTime := time.Now().UnixMilli() - r.startTime
				h.dispatch(resp, r.Request, useTime)
			}
		}
	}
	return n, nil
}
