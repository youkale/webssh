package echogy

import (
	"net"
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		charset     string
		wantErr     bool
		checkResult func(string) bool
	}{
		{
			name:    "valid length and charset",
			length:  8,
			charset: AlphaNum,
			wantErr: false,
			checkResult: func(s string) bool {
				return len(s) == 8 && containsOnlyChars(s, AlphaNum)
			},
		},
		{
			name:    "zero length",
			length:  0,
			charset: AlphaNum,
			wantErr: true,
		},
		{
			name:    "negative length",
			length:  -1,
			charset: AlphaNum,
			wantErr: true,
		},
		{
			name:    "empty charset",
			length:  8,
			charset: "",
			wantErr: true,
		},
		{
			name:    "digits only",
			length:  10,
			charset: Digits,
			wantErr: false,
			checkResult: func(s string) bool {
				return len(s) == 10 && containsOnlyChars(s, Digits)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateRandomString(tt.length, tt.charset)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateRandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.checkResult(got) {
				t.Errorf("generateRandomString() = %v, failed validation", got)
			}
		})
	}

	// Test randomness
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result, err := generateRandomString(8, AlphaNum)
		if err != nil {
			t.Errorf("generateRandomString() unexpected error = %v", err)
		}
		if results[result] {
			t.Error("generateRandomString() generated duplicate string")
		}
		results[result] = true
	}
}

func TestGenerateAccessId(t *testing.T) {
	tests := []struct {
		name    string
		raddr   net.Addr
		wantLen int
	}{
		{
			name: "TCP address",
			raddr: &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8080,
			},
			wantLen: 8,
		},
		{
			name: "Different TCP address",
			raddr: &net.TCPAddr{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 22,
			},
			wantLen: 8,
		},
		{
			name: "UDP address",
			raddr: &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8080,
			},
			wantLen: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := withAddrGenerateAccessId(tt.raddr)
			if err != nil {
				t.Errorf("withAddrGenerateAccessId() error = %v", err)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("withAddrGenerateAccessId() got length = %v, want %v", len(got), tt.wantLen)
			}
			if !containsOnlyChars(got, AlphaNum) {
				t.Errorf("withAddrGenerateAccessId() got = %v, contains invalid characters", got)
			}
		})
	}

	// Test caching for TCP addresses
	tcpAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	first, _ := withAddrGenerateAccessId(tcpAddr)
	second, _ := withAddrGenerateAccessId(tcpAddr)
	if first != second {
		t.Error("withAddrGenerateAccessId() cache not working, got different values for same TCP address")
	}
}

func TestWithIPGenerateAccessId(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")
	id, err := withIPGenerateAccessId(ip)
	if err != nil {
		t.Errorf("withIPGenerateAccessId() error = %v", err)
	}
	if len(id) != 8 {
		t.Errorf("withIPGenerateAccessId() got length = %v, want 8", len(id))
	}
}

func TestWithAddrGenerateAccessId(t *testing.T) {
	addr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	id, err := withAddrGenerateAccessId(addr)
	if err != nil {
		t.Errorf("withAddrGenerateAccessId() error = %v", err)
	}
	if len(id) != 8 {
		t.Errorf("withAddrGenerateAccessId() got length = %v, want 8", len(id))
	}
}

func TestParseHostAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		wantHost string
		wantPort uint32
		wantErr  bool
	}{
		{
			name:     "valid address with port",
			addr:     "localhost:8080",
			wantHost: "localhost",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "valid IP with port",
			addr:     "127.0.0.1:22",
			wantHost: "127.0.0.1",
			wantPort: 22,
			wantErr:  false,
		},
		{
			name:    "missing port",
			addr:    "localhost",
			wantErr: true,
		},
		{
			name:    "invalid port",
			addr:    "localhost:abc",
			wantErr: true,
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:     "IPv6 address with port",
			addr:     "[::1]:8080",
			wantHost: "::1",
			wantPort: 8080,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, err := parseHostAddr(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHostAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotHost != tt.wantHost {
					t.Errorf("parseHostAddr() gotHost = %v, want %v", gotHost, tt.wantHost)
				}
				if gotPort != tt.wantPort {
					t.Errorf("parseHostAddr() gotPort = %v, want %v", gotPort, tt.wantPort)
				}
			}
		})
	}
}

func TestParseHostAddrSimple(t *testing.T) {
	addr := "127.0.0.1:8080"
	host, port, err := parseHostAddr(addr)
	if err != nil {
		t.Errorf("parseHostAddr() error = %v", err)
	}
	if host != "127.0.0.1" {
		t.Errorf("parseHostAddr() got host = %v, want 127.0.0.1", host)
	}
	if port != 8080 {
		t.Errorf("parseHostAddr() got port = %v, want 8080", port)
	}
}

// Helper function to check if a string contains only characters from a given set
func containsOnlyChars(s, chars string) bool {
	for _, c := range s {
		if !strings.ContainsRune(chars, c) {
			return false
		}
	}
	return true
}
