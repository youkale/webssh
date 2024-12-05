package echogy

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"net"
	"time"
)

// bufferedReader implements an io.Reader that buffers data and supports re-reading
type bufferedReader struct {
	reader   io.Reader
	buffer   []byte
	position int
}

// newBufferedReader creates a new bufferedReader
func newBufferedReader(reader io.Reader) *bufferedReader {
	return &bufferedReader{
		reader:   reader,
		buffer:   make([]byte, 0),
		position: 0,
	}
}

type bufferedConn struct {
	reader io.Reader
	net.Conn
}

func (c *bufferedConn) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (b *bufferedReader) toBufferedConn(conn net.Conn) net.Conn {
	b.Reset()
	return &bufferedConn{
		reader: io.MultiReader(b, conn),
		Conn:   conn,
	}
}

// Read implements io.Reader interface
func (b *bufferedReader) Read(p []byte) (n int, err error) {
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
func (b *bufferedReader) Reset() {
	b.position = 0
}

// Seek sets the read position to a specific offset
func (b *bufferedReader) Seek(offset int64, whence int) (int64, error) {
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
func (b *bufferedReader) Bytes() []byte {
	return append([]byte(nil), b.buffer...)
}

// Len returns the total length of buffered data
func (b *bufferedReader) Len() int {
	return len(b.buffer)
}

// Position returns the current read position
func (b *bufferedReader) Position() int {
	return b.position
}

type wrappedConn struct {
	session ssh.Session
	gossh.Channel
}

func wrapChannelConn(session ssh.Session, channel gossh.Channel) *wrappedConn {
	return &wrappedConn{session: session, Channel: channel}
}

func (w *wrappedConn) bufferedReader() *bufferedReader {
	return newBufferedReader(w)
}

func (w *wrappedConn) LocalAddr() net.Addr {
	return w.session.LocalAddr()
}

func (w *wrappedConn) RemoteAddr() net.Addr {
	return w.session.RemoteAddr()
}

func (w *wrappedConn) SetDeadline(t time.Time) error {
	return nil
}

func (w *wrappedConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (w *wrappedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (w *wrappedConn) Close() error {
	return w.Channel.Close()
}
