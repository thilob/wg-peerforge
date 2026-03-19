package usecases

import (
	"github.com/thilob/wg-peerforge/internal/domain/models"
	"github.com/thilob/wg-peerforge/internal/infrastructure/renderers"
)

type RenderedServerConfig struct {
	ServerID string
	Target   string
	Content  string
}

func RenderServerConfig(server models.Server, peers []models.Peer) RenderedServerConfig {
	return RenderedServerConfig{
		ServerID: server.ID,
		Target:   "wg-quick",
		Content:  renderers.RenderWGQuickServerConfig(server, peers),
	}
}
