package echogy

import (
	"crypto/rand"
	"fmt"
	"github.com/karlseguin/ccache/v3"
	"net"
	"strconv"
	"time"
)

const (
	// Lowercase Character sets for random string generation
	Lowercase = "abcdefghijklmnopqrstuvwxyz"
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digits    = "0123456789"
	// AlphaNum Predefined character sets
	AlphaNum = Lowercase + Digits
)

// GenerateRandomString generates a random string of specified length using the given character set
func generateRandomString(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	if len(charset) == 0 {
		return "", fmt.Errorf("charset cannot be empty")
	}

	// Create a byte slice to store the result
	result := make([]byte, length)

	// Calculate the number of random bytes needed
	// We need 1 byte of randomness for each character in the result
	randomBytes := make([]byte, length)

	// Read random bytes
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Convert random bytes to characters from charset
	charsetLength := len(charset)
	for i := 0; i < length; i++ {
		// Use modulo to map random byte to charset index
		// This ensures uniform distribution
		result[i] = charset[randomBytes[i]%byte(charsetLength)]
	}

	return string(result), nil
}

var cache = ccache.New(ccache.Configure[string]().MaxSize(1_000_000))

func withIPGenerateAccessId(ip net.IP) (string, error) {
	cacheKey := ip.String()
	fetch, err := cache.Fetch(cacheKey, time.Hour*12, func() (string, error) {
		return generateRandomString(8, AlphaNum)
	})
	if nil != fetch && nil == err {
		return fetch.Value(), nil
	} else {
		return generateRandomString(8, AlphaNum)
	}
}

func withAddrGenerateAccessId(raddr net.Addr) (string, error) {
	switch raddr.(type) {
	case *net.TCPAddr:
		addr := raddr.(*net.TCPAddr)
		return withIPGenerateAccessId(addr.IP)
	case *net.UDPAddr:
		addr := raddr.(*net.UDPAddr)
		return withIPGenerateAccessId(addr.IP)
	default:
		return "", fmt.Errorf("unknown address type: %T", raddr)
	}
}

func generateAccessId() (string, error) {
	return generateRandomString(8, AlphaNum)
}

func parseHostAddr(addr string) (string, uint32, error) {
	host, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, err
	}
	return host, uint32(port), nil
}
