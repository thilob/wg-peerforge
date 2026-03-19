package renderers

import (
	"fmt"
	"strings"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

func RenderWGQuickServerConfig(server models.Server, peers []models.Peer) string {
	lines := []string{
		"[Interface]",
		renderAddressLine(server),
		fmt.Sprintf("ListenPort = %d", server.ListenPort),
		fmt.Sprintf("PrivateKey = <resolved:%s>", server.PrivateKeyRef),
	}

	if len(server.DNS) > 0 {
		lines = append(lines, fmt.Sprintf("DNS = %s", strings.Join(server.DNS, ", ")))
	}

	for _, peer := range peers {
		if peer.Status != models.PeerStatusActive {
			continue
		}
		lines = append(lines,
			"",
			"[Peer]",
			fmt.Sprintf("# %s", peer.Name),
			fmt.Sprintf("PublicKey = %s", peer.PublicKey),
		)
		if peer.PresharedKeyRef != "" {
			lines = append(lines, fmt.Sprintf("PresharedKey = <resolved:%s>", peer.PresharedKeyRef))
		}
		lines = append(lines, fmt.Sprintf("AllowedIPs = %s", renderPeerIPs(peer)))
	}

	return strings.Join(lines, "\n") + "\n"
}

func renderAddressLine(server models.Server) string {
	if server.AddressV6 == "" {
		return "Address = " + server.AddressV4
	}
	return fmt.Sprintf("Address = %s, %s", server.AddressV4, server.AddressV6)
}

func renderPeerIPs(peer models.Peer) string {
	if peer.IPv6 == "" {
		return peer.IPv4
	}
	return fmt.Sprintf("%s, %s", peer.IPv4, peer.IPv6)
}
