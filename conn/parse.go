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
	LocalIp    string
	LocalPort  string
	RemoteIp   string
	RemotePort string
	State      string
	Inode      string

	Proc string
	PID  string
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

// parse splits "IPHEX:PORTHEX" (e.g. "0100007F:1F90") into IP/port.
// IPHEX: 4-byte big-endian IPv4.
func parse(hex string) (string, string) {
	parts := strings.Split(hex, ":")
	iphex, porthex := parts[0], parts[1]

	ip := getip(iphex)

	port, _ := strconv.ParseInt(porthex, 16, 64)

	return ip, fmt.Sprintf("%d", port)
}

// getip decodes IPv4 hex bytes, reverses to host byte order, formats as dotted quad.
func getip(ip string) string {
	h, _ := hex.DecodeString(ip)
	for i, j := 0, len(h)-1; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}

	return net.IP(h).String()
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
