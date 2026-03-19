package ports

import "github.com/thilob/wg-peerforge/internal/domain/models"

type ConfigStore interface {
	SaveServer(server models.Server) error
	SavePeer(peer models.Peer) error
	ListServers() ([]models.Server, error)
	ListPeers(serverID string) ([]models.Peer, error)
}
