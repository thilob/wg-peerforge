package configstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

type FileStore struct {
	baseDir string
}

func NewFileStore(baseDir string) FileStore {
	if baseDir == "" {
		baseDir = "."
	}

	return FileStore{baseDir: baseDir}
}

func (s FileStore) SaveServer(server models.Server) error {
	return s.writeJSON(ServerConfigPath(server.ID), server)
}

func (s FileStore) SavePeer(peer models.Peer) error {
	return s.writeJSON(PeerConfigPath(peer.ServerID, peer.ID), peer)
}

func (s FileStore) ListServers() ([]models.Server, error) {
	serverRoot := s.path(filepath.Join("data", "servers"))
	entries, err := os.ReadDir(serverRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Server{}, nil
		}
		return nil, fmt.Errorf("read server directory: %w", err)
	}

	servers := make([]models.Server, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var server models.Server
		if err := s.readJSON(ServerConfigPath(entry.Name()), &server); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].ID < servers[j].ID
	})

	return servers, nil
}

func (s FileStore) ListPeers(serverID string) ([]models.Peer, error) {
	peerRoot := s.path(PeersDirectory(serverID))
	entries, err := os.ReadDir(peerRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Peer{}, nil
		}
		return nil, fmt.Errorf("read peer directory: %w", err)
	}

	peers := make([]models.Peer, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var peer models.Peer
		if err := s.readJSON(filepath.Join(PeersDirectory(serverID), entry.Name()), &peer); err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}

	sort.Slice(peers, func(i, j int) bool {
		return peers[i].ID < peers[j].ID
	})

	return peers, nil
}

func (s FileStore) readJSON(relativePath string, target any) error {
	file, err := os.Open(s.path(relativePath))
	if err != nil {
		return fmt.Errorf("open %s: %w", relativePath, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", relativePath, err)
	}

	return nil
}

func (s FileStore) writeJSON(relativePath string, value any) error {
	fullPath := s.path(relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", relativePath, err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", relativePath, err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		file.Close()
		return fmt.Errorf("encode %s: %w", relativePath, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", relativePath, err)
	}

	return nil
}

func (s FileStore) path(relativePath string) string {
	return filepath.Join(s.baseDir, relativePath)
}
