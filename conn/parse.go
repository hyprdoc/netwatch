package conn

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Connection struct {
	LocalIp    string
	LocalPort  string
	RemoteIp   string
	RemotePort string
	State      string
	Inode      string
}

func FetchConnections(path string) ([]Connection, error) {

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("error :", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var conns []Connection

	scanner.Scan()

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		localIp, localPort := parse(fields[1])
		remoteIp, remotePort := parse(fields[2])

		state := tcpState(fields[3])
		inode := fields[9]

		conns = append(conns, Connection{
			LocalIp:    localIp,
			LocalPort:  localPort,
			RemoteIp:   remoteIp,
			RemotePort: remotePort,
			State:      state,
			Inode:      inode,
		})

	}
	return conns, nil
}

// sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
// "0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345"

func parse(hex string) (string, string) {
	parts := strings.Split(hex, ":")
	iphex, porthex := parts[0], parts[1]

	ip := getip(iphex)

	port, _ := strconv.ParseInt(porthex, 16, 16)

	return ip, fmt.Sprintf("%d", port)
}

func getip(ip string) string {
	h, _ := hex.DecodeString(ip)
	for i, j := 0, len(h)-1; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}

	return net.IP(h).String()
}

func tcpState(code string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}

	if s, ok := states[code]; ok {
		return s
	}
	return code
}
