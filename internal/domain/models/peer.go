package models

type PeerStatus string

const (
	PeerStatusActive   PeerStatus = "active"
	PeerStatusDisabled PeerStatus = "disabled"
	PeerStatusExpired  PeerStatus = "expired"
)

type Peer struct {
	ID              string
	ServerID        string
	Name            string
	Status          PeerStatus
	IPv4            string
	IPv6            string
	PrivateKeyRef   string
	PublicKey       string
	PresharedKeyRef string
	AllowedIPs      []string
	DNS             []string
	Notes           string
	CreatedAt       string
	ExpiresAt       string
}
