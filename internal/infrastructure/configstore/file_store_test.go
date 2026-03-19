package configstore

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

func TestFileStoreSaveAndListServers(t *testing.T) {
	t.Parallel()

	store := NewFileStore(t.TempDir())
	server := models.Server{
		ID:                "alpha",
		Name:              "Alpha",
		InterfaceName:     "wg0",
		ListenPort:        51820,
		Endpoint:          "vpn.example.com",
		AddressV4:         "10.0.0.1/24",
		DNS:               []string{"1.1.1.1"},
		PrivateKeyRef:     "secret/server",
		PublicKey:         "public-server",
		DefaultAllowedIPs: []string{"0.0.0.0/0"},
		BackendMode:       models.RenderBackendWGQuick,
	}

	if err := store.SaveServer(server); err != nil {
		t.Fatalf("SaveServer() error = %v", err)
	}

	servers, err := store.ListServers()
	if err != nil {
		t.Fatalf("ListServers() error = %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("ListServers() length = %d, want 1", len(servers))
	}

	if !reflect.DeepEqual(servers[0], server) {
		t.Fatalf("ListServers() server = %#v, want %#v", servers[0], server)
	}
}

func TestFileStoreSaveAndListPeers(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	store := NewFileStore(baseDir)
	peerA := models.Peer{
		ID:            "peer-a",
		ServerID:      "alpha",
		Name:          "Peer A",
		Status:        models.PeerStatusActive,
		IPv4:          "10.0.0.2/32",
		PrivateKeyRef: "secret/peer-a",
		PublicKey:     "public-peer-a",
		AllowedIPs:    []string{"10.0.0.2/32"},
		DNS:           []string{"1.1.1.1"},
		CreatedAt:     "2026-03-19T12:00:00Z",
	}
	peerB := models.Peer{
		ID:            "peer-b",
		ServerID:      "alpha",
		Name:          "Peer B",
		Status:        models.PeerStatusDisabled,
		IPv4:          "10.0.0.3/32",
		PrivateKeyRef: "secret/peer-b",
		PublicKey:     "public-peer-b",
		AllowedIPs:    []string{"10.0.0.3/32"},
		CreatedAt:     "2026-03-19T12:05:00Z",
	}
	otherServerPeer := models.Peer{
		ID:            "peer-c",
		ServerID:      "beta",
		Name:          "Peer C",
		Status:        models.PeerStatusActive,
		IPv4:          "10.1.0.2/32",
		PrivateKeyRef: "secret/peer-c",
		PublicKey:     "public-peer-c",
		AllowedIPs:    []string{"10.1.0.2/32"},
		CreatedAt:     "2026-03-19T12:10:00Z",
	}

	for _, peer := range []models.Peer{peerB, otherServerPeer, peerA} {
		if err := store.SavePeer(peer); err != nil {
			t.Fatalf("SavePeer(%s) error = %v", peer.ID, err)
		}
	}

	peers, err := store.ListPeers("alpha")
	if err != nil {
		t.Fatalf("ListPeers() error = %v", err)
	}

	want := []models.Peer{peerA, peerB}
	if !reflect.DeepEqual(peers, want) {
		t.Fatalf("ListPeers() peers = %#v, want %#v", peers, want)
	}

	peerPath := filepath.Join(baseDir, PeerConfigPath("alpha", "peer-a"))
	if _, err := filepath.Abs(peerPath); err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
}
