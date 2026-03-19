package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type ServerSummary struct {
	ID            string
	Name          string
	InterfaceName string
	Endpoint      string
	AddressV4     string
	ListenPort    int
	BackendMode   string
}

type PeerSummary struct {
	ID        string
	Name      string
	Status    string
	IPv4      string
	Notes     string
	CreatedAt string
}

type DashboardData struct {
	DataRoot string
	Servers  []ServerView
}

type ServerView struct {
	Server ServerSummary
	Peers  []PeerSummary
}

type Dashboard struct {
	data        DashboardData
	selectedSrv int
	selectedPr  int
	focus       focusArea
}

type focusArea int

const (
	focusServers focusArea = iota
	focusPeers
)

type Key int

const (
	KeyUnknown Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyTab
	KeyQuit
	KeyRefresh
)

func Placeholder() string {
	return "Interactive prototype TUI is available via `wg-peerforge tui`."
}

func NewDashboard(data DashboardData) *Dashboard {
	d := &Dashboard{data: normalize(data)}
	if len(d.data.Servers) == 0 {
		d.focus = focusServers
		return d
	}
	if len(d.data.Servers[0].Peers) > 0 {
		d.focus = focusPeers
	}
	return d
}

func (d *Dashboard) Render() string {
	leftWidth := 34
	var lines []string
	lines = append(lines, ansiClear)
	lines = append(lines, "wg-peerforge  Prototype TUI")
	lines = append(lines, fmt.Sprintf("Data root: %s  Servers: %d  Peers: %d", d.data.DataRoot, len(d.data.Servers), d.totalPeers()))
	lines = append(lines, strings.Repeat("=", 100))

	left := d.renderServerList()
	right := d.renderDetailPane()
	maxLines := max(len(left), len(right))
	for i := 0; i < maxLines; i++ {
		l := lineAt(left, i)
		r := lineAt(right, i)
		lines = append(lines, padRight(l, leftWidth)+" | "+r)
	}

	lines = append(lines, strings.Repeat("-", 100))
	lines = append(lines, "Keys: Up/Down or j/k move  Left/Right or h/l switch pane  Tab switch pane  r refresh  q quit")
	lines = append(lines, "Current focus: "+d.focusLabel())
	return strings.Join(lines, "\n")
}

func (d *Dashboard) HandleKey(key Key) bool {
	switch key {
	case KeyQuit:
		return true
	case KeyTab:
		d.toggleFocus()
	case KeyLeft:
		d.focus = focusServers
	case KeyRight:
		if len(d.selectedServer().Peers) > 0 {
			d.focus = focusPeers
		}
	case KeyUp:
		d.moveSelection(-1)
	case KeyDown:
		d.moveSelection(1)
	}

	d.clampSelection()
	return false
}

func (d *Dashboard) SetData(data DashboardData) {
	currentServerID := ""
	currentPeerID := ""
	if server := d.selectedServer(); server.Server.ID != "" {
		currentServerID = server.Server.ID
	}
	if peer := d.selectedPeer(); peer.ID != "" {
		currentPeerID = peer.ID
	}

	d.data = normalize(data)
	d.selectedSrv = findServerIndex(d.data.Servers, currentServerID)
	d.selectedPr = findPeerIndex(d.selectedServer().Peers, currentPeerID)
	d.clampSelection()
}

func (d *Dashboard) renderServerList() []string {
	lines := []string{"Servers"}
	if len(d.data.Servers) == 0 {
		lines = append(lines, "", "  no servers stored yet", "", "  create one with:", "  wg-peerforge create-server -id alpha -name Alpha -endpoint vpn.example.com")
		return lines
	}

	for i, server := range d.data.Servers {
		cursor := " "
		if i == d.selectedSrv && d.focus == focusServers {
			cursor = ">"
		} else if i == d.selectedSrv {
			cursor = "*"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, server.Server.ID))
		lines = append(lines, fmt.Sprintf("  %s", truncate(server.Server.Name, 28)))
		lines = append(lines, fmt.Sprintf("  %s  %d peers", truncate(server.Server.InterfaceName, 10), len(server.Peers)))
		lines = append(lines, "")
	}

	return lines
}

func (d *Dashboard) renderDetailPane() []string {
	if len(d.data.Servers) == 0 {
		return []string{
			"Details",
			"",
			"No data yet.",
			"",
			"Once servers exist, this pane will show",
			"server details and the peer list for the current selection.",
		}
	}

	server := d.selectedServer()
	lines := []string{
		"Details",
		"",
		"Server",
		fmt.Sprintf("  ID: %s", server.Server.ID),
		fmt.Sprintf("  Name: %s", server.Server.Name),
		fmt.Sprintf("  Interface: %s", server.Server.InterfaceName),
		fmt.Sprintf("  Endpoint: %s", server.Server.Endpoint),
		fmt.Sprintf("  Tunnel IPv4: %s", server.Server.AddressV4),
		fmt.Sprintf("  Listen port: %d", server.Server.ListenPort),
		fmt.Sprintf("  Backend: %s", server.Server.BackendMode),
		"",
		"Peers",
	}

	if len(server.Peers) == 0 {
		lines = append(lines, "  none yet")
		return lines
	}

	for i, peer := range server.Peers {
		cursor := " "
		if i == d.selectedPr && d.focus == focusPeers {
			cursor = ">"
		} else if i == d.selectedPr {
			cursor = "*"
		}
		lines = append(lines, fmt.Sprintf("%s %s  %s  %s", cursor, peer.ID, peer.Status, peer.IPv4))
	}

	peer := d.selectedPeer()
	if peer.ID == "" {
		return lines
	}

	lines = append(lines,
		"",
		"Selected peer",
		fmt.Sprintf("  ID: %s", peer.ID),
		fmt.Sprintf("  Name: %s", peer.Name),
		fmt.Sprintf("  Status: %s", peer.Status),
		fmt.Sprintf("  IPv4: %s", peer.IPv4),
		fmt.Sprintf("  Created: %s", blankFallback(peer.CreatedAt, "-")),
		fmt.Sprintf("  Notes: %s", blankFallback(peer.Notes, "-")),
	)
	return lines
}

func (d *Dashboard) selectedServer() ServerView {
	if len(d.data.Servers) == 0 || d.selectedSrv < 0 || d.selectedSrv >= len(d.data.Servers) {
		return ServerView{}
	}
	return d.data.Servers[d.selectedSrv]
}

func (d *Dashboard) selectedPeer() PeerSummary {
	server := d.selectedServer()
	if len(server.Peers) == 0 || d.selectedPr < 0 || d.selectedPr >= len(server.Peers) {
		return PeerSummary{}
	}
	return server.Peers[d.selectedPr]
}

func (d *Dashboard) moveSelection(delta int) {
	if d.focus == focusPeers {
		d.selectedPr += delta
		return
	}
	d.selectedSrv += delta
}

func (d *Dashboard) toggleFocus() {
	if len(d.data.Servers) == 0 {
		d.focus = focusServers
		return
	}
	if d.focus == focusServers && len(d.selectedServer().Peers) > 0 {
		d.focus = focusPeers
		return
	}
	d.focus = focusServers
}

func (d *Dashboard) clampSelection() {
	if len(d.data.Servers) == 0 {
		d.selectedSrv = 0
		d.selectedPr = 0
		d.focus = focusServers
		return
	}

	if d.selectedSrv < 0 {
		d.selectedSrv = 0
	}
	if d.selectedSrv >= len(d.data.Servers) {
		d.selectedSrv = len(d.data.Servers) - 1
	}

	peers := d.selectedServer().Peers
	if len(peers) == 0 {
		d.selectedPr = 0
		if d.focus == focusPeers {
			d.focus = focusServers
		}
		return
	}

	if d.selectedPr < 0 {
		d.selectedPr = 0
	}
	if d.selectedPr >= len(peers) {
		d.selectedPr = len(peers) - 1
	}
}

func (d *Dashboard) totalPeers() int {
	total := 0
	for _, server := range d.data.Servers {
		total += len(server.Peers)
	}
	return total
}

func (d *Dashboard) focusLabel() string {
	if d.focus == focusPeers {
		return "peers"
	}
	return "servers"
}

func normalize(data DashboardData) DashboardData {
	servers := append([]ServerView(nil), data.Servers...)
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Server.ID < servers[j].Server.ID
	})
	for i := range servers {
		sort.Slice(servers[i].Peers, func(a, b int) bool {
			return servers[i].Peers[a].ID < servers[i].Peers[b].ID
		})
	}
	data.Servers = servers
	return data
}

func ReadKey(r io.Reader) (Key, error) {
	var buf [3]byte
	n, err := r.Read(buf[:1])
	if err != nil {
		return KeyUnknown, err
	}
	if n == 0 {
		return KeyUnknown, nil
	}

	switch buf[0] {
	case 'q', 'Q':
		return KeyQuit, nil
	case 'r', 'R':
		return KeyRefresh, nil
	case '\t':
		return KeyTab, nil
	case 'k', 'K':
		return KeyUp, nil
	case 'j', 'J':
		return KeyDown, nil
	case 'h', 'H':
		return KeyLeft, nil
	case 'l', 'L':
		return KeyRight, nil
	case 0x1b:
		if _, err := r.Read(buf[1:2]); err != nil {
			return KeyUnknown, nil
		}
		if buf[1] != '[' {
			return KeyUnknown, nil
		}
		if _, err := r.Read(buf[2:3]); err != nil {
			return KeyUnknown, nil
		}
		switch buf[2] {
		case 'A':
			return KeyUp, nil
		case 'B':
			return KeyDown, nil
		case 'C':
			return KeyRight, nil
		case 'D':
			return KeyLeft, nil
		}
	}

	return KeyUnknown, nil
}

const ansiClear = "\x1b[2J\x1b[H"

func findServerIndex(servers []ServerView, id string) int {
	for i, server := range servers {
		if server.Server.ID == id {
			return i
		}
	}
	return 0
}

func findPeerIndex(peers []PeerSummary, id string) int {
	for i, peer := range peers {
		if peer.ID == id {
			return i
		}
	}
	return 0
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func truncate(value string, width int) string {
	if len(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	return value[:width-1] + "…"
}

func blankFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
