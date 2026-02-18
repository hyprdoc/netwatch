package conn

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input        string
		expectedIP   string
		expectedPort string
	}{
		{"0100007F:1F90", "127.0.0.1", "8080"},
		{"00000000:0000", "0.0.0.0", "0"},
		{"01020304:0050", "4.3.2.1", "80"},
	}

	for _, tt := range tests {
		ip, port := parseHexAddr(tt.input)
		if ip != tt.expectedIP {
			t.Errorf("parse(%q) IP = %v, want %v", tt.input, ip, tt.expectedIP)
		}
		if port != tt.expectedPort {
			t.Errorf("parse(%q) Port = %v, want %v", tt.input, port, tt.expectedPort)
		}
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// ipv4
		{"0100007F", "127.0.0.1"},
		{"00000000", "0.0.0.0"},
		{"01020304", "4.3.2.1"},
		{"FFFFFFFF", "255.255.255.255"},
		// ipv6
		{input: "01000000000000000000000000000000", expected: "::1"},
		{input: "FFEEDDCCBBAA99887766554433221100", expected: "1011:2233:4455:6677:8899:aabb:ccdd:eeff"},
		{input: "01020304050607080910111213141516", expected: "1615:1413:1211:1009:0807:0605:0403:0201"},
		{input: "00000000000000000000000000000002", expected: "::2"},
		{input: "0000000000000000000000000000FFFF", expected: "::ffff"},
		{input: "7F0000FF0000000000000000FFFF0000", expected: "::ffff:127.0.0.1"},
	}

	for _, tt := range tests {
		result := decodeIP(tt.input)
		if result != tt.expected {
			t.Errorf("decodeIP(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTcpState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01", "ESTABLISHED"},
		{"02", "SYN_SENT"},
		{"03", "SYN_RECV"},
		{"04", "FIN_WAIT1"},
		{"05", "FIN_WAIT2"},
		{"06", "TIME_WAIT"},
		{"07", "CLOSE"},
		{"08", "CLOSE_WAIT"},
		{"09", "LAST_ACK"},
		{"0A", "LISTEN"},
		{"0B", "CLOSING"},
		{"99", "99"}, // Unknown state
	}

	for _, tt := range tests {
		result := tcpState(tt.input)
		if result != tt.expected {
			t.Errorf("tcpState(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
