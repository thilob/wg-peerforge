package models

type RenderBackend string

const (
	RenderBackendWGQuick        RenderBackend = "wg-quick"
	RenderBackendNetworkManager RenderBackend = "networkmanager"
	RenderBackendNetworkd       RenderBackend = "networkd"
	RenderBackendExportOnly     RenderBackend = "export-only"
)

type Server struct {
	ID                string
	Name              string
	InterfaceName     string
	ListenPort        int
	Endpoint          string
	AddressV4         string
	AddressV6         string
	DNS               []string
	PrivateKeyRef     string
	PublicKey         string
	DefaultAllowedIPs []string
	BackendMode       RenderBackend
}
