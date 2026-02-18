package conn

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Connection represents a TCP connection parsed from /proc/net/tcp.
// Fields from columns: local/remote addr, state, inode.
// Proc/PID matched post-parse via inode -> /proc/<PID>/fd socket symlinks.
type Connection struct {
	Protocol   string // tcp,tcp6,udp,udp6
	LocalIp    string // [IP_ADDRESS]
	LocalPort  string // 80
	RemoteIp   string // [IP_ADDRESS]
	RemotePort string // 54321
	State      string // ESTABLISHED, LISTEN, TIME_WAIT, etc.
	Inode      string // inode number

	Proc string // process name
	PID  string // process id
}

func FetchAllConnections() ([]Connection, error) {

	sources := []string{
		"/proc/net/tcp",
		"/proc/net/tcp6",
	}

	var conns []Connection

	for _, source := range sources {
		conn, err := FetchConnections(source)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn...)
	}

	return conns, nil
}

// FetchConnections parses active TCP sockets from netPath (e.g. "/proc/net/tcp").
// Enriches with owning process via inode matching in /proc/<PID>/fd.
// Linux-only; non-root users see only own procs (perm denied others).
// Optimizations: unique inodes, early loop exits on resolution.
func FetchConnections(netPath string) ([]Connection, error) {

	// Parse non-header lines: "sl local_address:port rem_address:port st ... uid timeout inode"
	// fetching all the live tcp connections
	file, err := os.Open(netPath)
	if err != nil {
		return nil, fmt.Errorf("[FetchConnections] failed to open [%s] error: %w", netPath, err)
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

		localIp, localPort := parseHexAddr(fields[1])
		remoteIp, remotePort := parseHexAddr(fields[2])

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

	// Build inode set for targeted matching; delete resolved inodes for early exits
	// fetching and mapping processes to inodes
	uniqueInodes := make(map[string]struct{}, len(conns))

	for _, c := range conns {
		uniqueInodes[c.Inode] = struct{}{}
	}

	// inode -> first-matching {pid, procname}
	inodeToProc := make(map[string]struct {
		pid  string
		proc string
	})

	// Iterate /proc/<PID> dirs (numeric names only)
	// read all the entries inside proc
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	// Track if any /proc/<PID>/fd access denied (label unowned conns later)
	permissionDenied := false

	for _, p := range procEntries {

		// Early exit: all target inodes matched
		// exit if all inodes are resolved
		if len(uniqueInodes) == 0 {
			break
		}

		pid := p.Name()

		// Quick PID check: skip non-numeric dir names (e.g. "self", "sys")
		// filter out the processes that does not start with a integer
		if pid[0] < '0' || pid[0] > '9' {
			continue
		}

		fdDir := filepath.Join("/proc", pid, "fd")

		fdEntries, err := os.ReadDir(fdDir)
		if err != nil {
			// Typically EACCES for other users' PIDs (non-root)
			// check if its permission issues
			if os.IsPermission(err) {
				permissionDenied = true
			}
			continue
		}

		commBytes, err := os.ReadFile(filepath.Join("/proc", pid, "comm"))
		if err != nil {
			continue
		}

		procName := strings.TrimSpace(string(commBytes))

		// Scan FD symlinks for "socket:[inode]" matching uniqueInodes
		for _, fd := range fdEntries {

			// Early exit per-PID
			if len(uniqueInodes) == 0 {
				break
			}
			fdPath := filepath.Join(fdDir, fd.Name())

			link, err := os.Readlink(fdPath)
			if err != nil {
				continue
			}

			if !strings.HasPrefix(link, "socket:[") {
				continue
			}

			inode := strings.TrimSuffix(
				strings.TrimPrefix(link, "socket:["),
				"]",
			)

			if _, ok := uniqueInodes[inode]; !ok {
				continue
			}

			inodeToProc[inode] = struct {
				pid  string
				proc string
			}{
				pid:  pid,
				proc: procName,
			}

			delete(uniqueInodes, inode)
		}

	}

	// Assign Proc/PID by inode; fallbacks for unmatched/perm denied
	for i := range conns {

		// Handle rare inode=0 (kernel reserved?)
		if conns[i].Inode == "0" {
			conns[i].Proc = "No Owner"
			continue
		}

		if p, ok := inodeToProc[conns[i].Inode]; ok {
			conns[i].PID = p.pid
			conns[i].Proc = p.proc
			continue
		}

		if permissionDenied {
			conns[i].Proc = "Permission Denied"
		} else {
			conns[i].Proc = "Unknown"
		}

	}

	return conns, nil
}

// sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
// "0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345"

// parseHexAddr parses "IPHEX:PORTHEX" into IP and port strings.
// Supports IPv4 (8 hex chars) and IPv6 (32 hex chars).
func parseHexAddr(s string) (string, string) {
	// find colon without allocation
	i := strings.LastIndexByte(s, ':')
	if i < 0 {
		return "", ""
	}

	iphex := s[:i]
	porthex := s[i+1:]

	ip := decodeIP(iphex)

	port, err := strconv.ParseUint(porthex, 16, 16)
	if err != nil {
		return ip, ""
	}

	return ip, strconv.FormatUint(port, 10)
}

// decodeIP handles both IPv4 and IPv6 hex encodings.
func decodeIP(iphex string) string {
	b, err := hex.DecodeString(iphex)
	if err != nil {
		return ""
	}
	switch len(b) {
	case net.IPv4len:
		// reverse for little-endian IPv4 (like /proc)
		return net.IPv4(b[3], b[2], b[1], b[0]).String()
	case net.IPv6len:
		// reverse for little-endian IPv6 (like /proc)
		for i := 0; i < len(b)/2; i++ {
			b[i], b[len(b)-1-i] = b[len(b)-1-i], b[i]
		}
		return net.IP(b).String()
	}
	return ""
}

// tcpState maps /proc/net/tcp 2-digit hex state codes to human names.
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
