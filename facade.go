package webssh

import (
	"bufio"
	"context"
	"fmt"
	"github.com/youkale/webssh/logger"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	NotFound = `HTTP/1.0 404 Not Found
Server: webs.sh
Content-Length: %d

Tunnel %s not found
`

	BadRequest = `HTTP/1.0 400 Bad Request
Server: webs.sh
Content-Length: 12

Bad Request
`
)

type repeatRead struct {
	preReader io.Reader
	net.Conn
}

func (re *repeatRead) Read(p []byte) (n int, err error) {
	return re.preReader.Read(p)
}

// BufferedReader implements an io.Reader that buffers data and supports re-reading
type BufferedReader struct {
	reader   io.Reader
	buffer   []byte
	position int
}

// NewBufferedReader creates a new BufferedReader
func NewBufferedReader(reader io.Reader) *BufferedReader {
	return &BufferedReader{
		reader:   reader,
		buffer:   make([]byte, 0),
		position: 0,
	}
}

// Read implements io.Reader interface
func (b *BufferedReader) Read(p []byte) (n int, err error) {
	// If we have buffered data and haven't reached the end
	if b.position < len(b.buffer) {
		n = copy(p, b.buffer[b.position:])
		b.position += n
		return n, nil
	}

	// Read new data from the underlying reader
	n, err = b.reader.Read(p)
	if n > 0 {
		// Append new data to our buffer
		b.buffer = append(b.buffer, p[:n]...)
		b.position += n
	}
	return n, err
}

// Reset resets the read position to the beginning of the buffer
func (b *BufferedReader) Reset() {
	b.position = 0
}

// Seek sets the read position to a specific offset
func (b *BufferedReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(b.position) + offset
	case io.SeekEnd:
		abs = int64(len(b.buffer)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if abs < 0 {
		return 0, fmt.Errorf("negative position: %d", abs)
	}

	if abs > int64(len(b.buffer)) {
		return 0, fmt.Errorf("seek position %d beyond buffer length %d", abs, len(b.buffer))
	}

	b.position = int(abs)
	return abs, nil
}

// Bytes returns a copy of the buffered data
func (b *BufferedReader) Bytes() []byte {
	return append([]byte(nil), b.buffer...)
}

// Len returns the total length of buffered data
func (b *BufferedReader) Len() int {
	return len(b.buffer)
}

// Position returns the current read position
func (b *BufferedReader) Position() int {
	return b.position
}

func badRequest(conn net.Conn) {
	conn.Write([]byte(BadRequest))
	conn.Close()
}

func notFound(id string, conn net.Conn) {
	conn.Write([]byte(fmt.Sprintf(NotFound, len(id)+18, id)))
	conn.Close()
}

func facadeServe(ctx context.Context, addr string, forward func(facadeId string, conn net.Conn) bool) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("start Listen", err, map[string]interface{}{
			"module":  "facade",
			"address": addr,
		})
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := ln.Accept()
			if nil != err {
				logger.Error("start Accept", err, map[string]interface{}{
					"module":  "facade",
					"address": addr,
				})
				c.Close()
			} else {
				go func() {
					reader := NewBufferedReader(c)
					req, err := http.ReadRequest(bufio.NewReader(reader))
					if err != nil {
						logger.Warn("bad request", map[string]interface{}{
							"module": "facade",
						})
						badRequest(c)
						return
					}

					domainSep := strings.Split(req.Host, ".")
					if len(domainSep) <= 1 {
						logger.Warn("bad request", map[string]interface{}{
							"module": "facade",
							"method": req.Method,
							"path":   req.URL.Path,
							"host":   req.Host,
						})
						badRequest(c)
						return
					}
					id := domainSep[0]
					reader.Reset()

					canForward := forward(id, &repeatRead{
						preReader: io.MultiReader(reader, c),
						Conn:      c,
					})

					if canForward {
						logger.Debug("found forward", map[string]interface{}{
							"module":   "facade",
							"method":   req.Method,
							"accessId": id,
							"path":     req.URL.Path,
						})
					} else {
						notFound(id, c)
						logger.Warn("not found forward", map[string]interface{}{
							"module":   "facade",
							"method":   req.Method,
							"accessId": id,
							"path":     req.URL.Path,
						})
						return
					}
				}()
			}
		}
	}
}
