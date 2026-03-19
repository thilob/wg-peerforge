package validation

import (
	"strings"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

func ValidatePeer(peer models.Peer) Result {
	errors := []string{}

	if strings.TrimSpace(peer.Name) == "" {
		errors = append(errors, "peer name is required")
	}
	if !strings.Contains(peer.IPv4, "/") {
		errors = append(errors, "peer IPv4 address must use CIDR notation")
	}
	if strings.TrimSpace(peer.PrivateKeyRef) == "" {
		errors = append(errors, "peer private key reference is required")
	}
	if strings.TrimSpace(peer.PublicKey) == "" {
		errors = append(errors, "peer public key is required")
	}
	if len(peer.AllowedIPs) == 0 {
		errors = append(errors, "peer must define at least one AllowedIPs entry")
	}

	return Result{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}
