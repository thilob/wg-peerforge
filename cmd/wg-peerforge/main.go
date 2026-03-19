package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/thilob/wg-peerforge/internal/app/ports"
	"github.com/thilob/wg-peerforge/internal/app/usecases"
	"github.com/thilob/wg-peerforge/internal/domain/models"
	"github.com/thilob/wg-peerforge/internal/infrastructure/configstore"
	"github.com/thilob/wg-peerforge/internal/infrastructure/keys"
	"github.com/thilob/wg-peerforge/internal/ui"
)

type application struct {
	store      ports.ConfigStore
	keyService ports.KeyService
}

func main() {
	app := application{
		store:      configstore.NewFileStore("."),
		keyService: keys.NewLocalService(),
	}

	if err := app.run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func (app application) run(args []string) error {
	if len(args) == 0 {
		printBanner()
		printUsage()
		return nil
	}

	switch args[0] {
	case "create-server":
		return app.runCreateServer(args[1:])
	case "create-peer":
		return app.runCreatePeer(args[1:])
	case "tui":
		return app.runTUI(args[1:])
	case "list-servers":
		return app.runListServers(args[1:])
	case "list-peers":
		return app.runListPeers(args[1:])
	case "render-server":
		return app.runRenderServer(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printBanner() {
	fmt.Println("wg-peerforge")
	fmt.Println("Server and peer manager foundation")
	fmt.Println()
}

func printUsage() {
	fmt.Println(ui.Placeholder())
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  wg-peerforge create-server [flags]")
	fmt.Println("  wg-peerforge create-peer [flags]")
	fmt.Println("  wg-peerforge tui")
	fmt.Println("  wg-peerforge list-servers")
	fmt.Println("  wg-peerforge list-peers -server-id <id>")
	fmt.Println("  wg-peerforge render-server -server-id <id>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  wg-peerforge create-server -id alpha -name Alpha -endpoint vpn.example.com")
	fmt.Println("  wg-peerforge create-peer -server-id alpha -id phone -name \"Phone\"")
	fmt.Println("  wg-peerforge tui")
}

func (app application) runCreateServer(args []string) error {
	fs := flag.NewFlagSet("create-server", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.String("id", "", "Stable server identifier")
	name := fs.String("name", "", "Display name")
	iface := fs.String("interface", "wg0", "WireGuard interface name")
	port := fs.Int("port", 51820, "Listen port")
	endpoint := fs.String("endpoint", "", "Public endpoint or hostname")
	addressV4 := fs.String("address-v4", "10.0.0.1/24", "Tunnel IPv4 CIDR")
	addressV6 := fs.String("address-v6", "", "Optional tunnel IPv6 CIDR")
	dns := fs.String("dns", "", "Comma-separated DNS servers")
	allowedIPs := fs.String("default-allowed-ips", "0.0.0.0/0", "Comma-separated default AllowedIPs")
	backend := fs.String("backend", string(models.RenderBackendWGQuick), "Render backend")

	if err := fs.Parse(args); err != nil {
		return err
	}

	server, err := usecases.CreateServer(usecases.CreateServerInput{
		ID:                *id,
		Name:              *name,
		InterfaceName:     *iface,
		ListenPort:        *port,
		Endpoint:          *endpoint,
		AddressV4:         *addressV4,
		AddressV6:         *addressV6,
		DNS:               splitCSV(*dns),
		DefaultAllowedIPs: splitCSV(*allowedIPs),
		BackendMode:       models.RenderBackend(*backend),
	}, struct {
		ConfigStore ports.ConfigStore
		KeyService  ports.KeyService
	}{
		ConfigStore: app.store,
		KeyService:  app.keyService,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Created server %s (%s)\n", server.ID, server.Name)
	fmt.Printf("Stored at %s\n", configstore.ServerConfigPath(server.ID))
	return nil
}

func (app application) runCreatePeer(args []string) error {
	fs := flag.NewFlagSet("create-peer", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	serverID := fs.String("server-id", "", "Server identifier")
	id := fs.String("id", "", "Stable peer identifier")
	name := fs.String("name", "", "Display name")
	allowedIPs := fs.String("allowed-ips", "", "Comma-separated AllowedIPs")
	dns := fs.String("dns", "", "Comma-separated DNS servers")
	notes := fs.String("notes", "", "Optional notes")

	if err := fs.Parse(args); err != nil {
		return err
	}

	server, err := app.findServer(*serverID)
	if err != nil {
		return err
	}

	peers, err := app.store.ListPeers(server.ID)
	if err != nil {
		return fmt.Errorf("load existing peers: %w", err)
	}

	peer, err := usecases.CreatePeer(usecases.CreatePeerInput{
		ID:            *id,
		ServerID:      server.ID,
		Name:          *name,
		PoolCIDR:      server.AddressV4,
		AllocatedIPv4: allocatedPeerIPv4s(server, peers),
		AllowedIPs:    splitCSV(*allowedIPs),
		DNS:           splitCSV(*dns),
		Notes:         *notes,
	}, struct {
		ConfigStore ports.ConfigStore
		KeyService  ports.KeyService
	}{
		ConfigStore: app.store,
		KeyService:  app.keyService,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Created peer %s (%s) for server %s\n", peer.ID, peer.Name, peer.ServerID)
	fmt.Printf("Assigned IPv4 %s\n", peer.IPv4)
	fmt.Printf("Stored at %s\n", configstore.PeerConfigPath(peer.ServerID, peer.ID))
	return nil
}

func (app application) runListServers(args []string) error {
	fs := flag.NewFlagSet("list-servers", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	servers, err := app.store.ListServers()
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		fmt.Println("No servers stored yet.")
		return nil
	}

	for _, server := range servers {
		fmt.Printf("%s\t%s\t%s\t%d\n", server.ID, server.Name, server.InterfaceName, server.ListenPort)
	}

	return nil
}

func (app application) runTUI(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	data, err := app.loadDashboardData()
	if err != nil {
		return err
	}

	dashboard := ui.NewDashboard(data)

	restore, err := enableRawMode(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("enable raw mode: %w", err)
	}
	defer restore()

	for {
		fmt.Print(dashboard.Render())

		key, err := ui.ReadKey(os.Stdin)
		if err != nil {
			return fmt.Errorf("read key: %w", err)
		}

		if key == ui.KeyRefresh {
			data, err := app.loadDashboardData()
			if err != nil {
				return err
			}
			dashboard.SetData(data)
			continue
		}

		if quit := dashboard.HandleKey(key); quit {
			fmt.Print("\x1b[2J\x1b[H")
			return nil
		}
	}
}

func (app application) runListPeers(args []string) error {
	fs := flag.NewFlagSet("list-peers", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	serverID := fs.String("server-id", "", "Server identifier")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*serverID) == "" {
		return errors.New("server-id is required")
	}

	peers, err := app.store.ListPeers(*serverID)
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		fmt.Printf("No peers stored for server %s.\n", *serverID)
		return nil
	}

	for _, peer := range peers {
		fmt.Printf("%s\t%s\t%s\t%s\n", peer.ID, peer.Name, peer.Status, peer.IPv4)
	}

	return nil
}

func (app application) runRenderServer(args []string) error {
	fs := flag.NewFlagSet("render-server", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	serverID := fs.String("server-id", "", "Server identifier")
	writeFile := fs.Bool("write", true, "Write rendered config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	server, err := app.findServer(*serverID)
	if err != nil {
		return err
	}

	peers, err := app.store.ListPeers(server.ID)
	if err != nil {
		return fmt.Errorf("load peers: %w", err)
	}

	rendered := usecases.RenderServerConfig(server, peers)
	if *writeFile {
		targetPath := filepath.Join(".", configstore.RenderedServerConfigPath(server.ID, server.InterfaceName))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("create render directory: %w", err)
		}
		if err := os.WriteFile(targetPath, []byte(rendered.Content), 0o600); err != nil {
			return fmt.Errorf("write rendered config: %w", err)
		}
		fmt.Printf("Rendered config written to %s\n", configstore.RenderedServerConfigPath(server.ID, server.InterfaceName))
		return nil
	}

	fmt.Print(rendered.Content)
	return nil
}

func (app application) findServer(serverID string) (models.Server, error) {
	if strings.TrimSpace(serverID) == "" {
		return models.Server{}, errors.New("server-id is required")
	}

	servers, err := app.store.ListServers()
	if err != nil {
		return models.Server{}, fmt.Errorf("load servers: %w", err)
	}

	for _, server := range servers {
		if server.ID == serverID {
			return server, nil
		}
	}

	return models.Server{}, fmt.Errorf("server %q not found", serverID)
}

func allocatedPeerIPv4s(server models.Server, peers []models.Peer) []string {
	addresses := make([]string, 0, len(peers)+1)
	if server.AddressV4 != "" {
		addresses = append(addresses, server.AddressV4)
	}

	for _, peer := range peers {
		if peer.IPv4 == "" {
			continue
		}
		addresses = append(addresses, peer.IPv4)
	}
	return addresses
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

func (app application) loadDashboardData() (ui.DashboardData, error) {
	servers, err := app.store.ListServers()
	if err != nil {
		return ui.DashboardData{}, fmt.Errorf("load servers: %w", err)
	}

	views := make([]ui.ServerView, 0, len(servers))
	for _, server := range servers {
		peers, err := app.store.ListPeers(server.ID)
		if err != nil {
			return ui.DashboardData{}, fmt.Errorf("load peers for %s: %w", server.ID, err)
		}

		view := ui.ServerView{
			Server: ui.ServerSummary{
				ID:            server.ID,
				Name:          server.Name,
				InterfaceName: server.InterfaceName,
				Endpoint:      server.Endpoint,
				AddressV4:     server.AddressV4,
				ListenPort:    server.ListenPort,
				BackendMode:   string(server.BackendMode),
			},
			Peers: make([]ui.PeerSummary, 0, len(peers)),
		}

		for _, peer := range peers {
			view.Peers = append(view.Peers, ui.PeerSummary{
				ID:        peer.ID,
				Name:      peer.Name,
				Status:    string(peer.Status),
				IPv4:      peer.IPv4,
				Notes:     peer.Notes,
				CreatedAt: peer.CreatedAt,
			})
		}

		views = append(views, view)
	}

	return ui.DashboardData{
		DataRoot: "./data",
		Servers:  views,
	}, nil
}

func enableRawMode(fd uintptr) (func(), error) {
	original, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	raw := *original
	raw.Iflag &^= syscall.ICRNL | syscall.INLCR | syscall.IGNCR | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := setTermios(fd, &raw); err != nil {
		return nil, err
	}

	return func() {
		_ = setTermios(fd, original)
	}, nil
}

func getTermios(fd uintptr) (*syscall.Termios, error) {
	termios := &syscall.Termios{}
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return termios, nil
}

func setTermios(fd uintptr, termios *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
