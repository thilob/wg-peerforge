package ports

import "github.com/thilob/wg-peerforge/internal/domain/models"

type BackendDetection struct {
	Mode   models.RenderBackend
	Detail string
}

type BackendAdapter interface {
	Detect() (BackendDetection, error)
	ApplyServerConfig(serverID string) error
}
