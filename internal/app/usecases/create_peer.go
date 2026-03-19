package usecases

import (
	"fmt"
	"strings"
	"time"

	"github.com/thilob/wg-peerforge/internal/app/ports"
	"github.com/thilob/wg-peerforge/internal/domain/models"
	"github.com/thilob/wg-peerforge/internal/domain/services"
	"github.com/thilob/wg-peerforge/internal/domain/validation"
)

type CreatePeerInput struct {
	ID            string
	ServerID      string
	Name          string
	PoolCIDR      string
	AllocatedIPv4 []string
	AllowedIPs    []string
	DNS           []string
	Notes         string
}

func CreatePeer(input CreatePeerInput, deps struct {
	ConfigStore ports.ConfigStore
	KeyService  ports.KeyService
}) (models.Peer, error) {
	keyPair, err := deps.KeyService.GenerateKeyPair("peer:" + input.ServerID + ":" + input.ID)
	if err != nil {
		return models.Peer{}, fmt.Errorf("generate peer key pair: %w", err)
	}

	ipv4, err := services.NextAvailableIPv4(models.AddressPool{
		CIDR:      input.PoolCIDR,
		Allocated: input.AllocatedIPv4,
	})
	if err != nil {
		return models.Peer{}, err
	}

	peer := models.Peer{
		ID:            input.ID,
		ServerID:      input.ServerID,
		Name:          input.Name,
		Status:        models.PeerStatusActive,
		IPv4:          ipv4,
		PrivateKeyRef: keyPair.PrivateKeyRef,
		PublicKey:     keyPair.PublicKey,
		AllowedIPs:    defaultAllowedIPs(input.AllowedIPs),
		DNS:           defaultStrings(input.DNS),
		Notes:         input.Notes,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	result := validation.ValidatePeer(peer)
	if !result.Valid {
		return models.Peer{}, fmt.Errorf(strings.Join(result.Errors, "; "))
	}

	if err := deps.ConfigStore.SavePeer(peer); err != nil {
		return models.Peer{}, fmt.Errorf("save peer: %w", err)
	}

	return peer, nil
}
