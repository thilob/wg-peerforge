package configstore

import "path/filepath"

func ServerDirectory(serverID string) string {
	return filepath.Join("data", "servers", serverID)
}

func ServerConfigPath(serverID string) string {
	return filepath.Join(ServerDirectory(serverID), "server.toml")
}

func PeersDirectory(serverID string) string {
	return filepath.Join(ServerDirectory(serverID), "peers")
}

func PeerConfigPath(serverID, peerID string) string {
	return filepath.Join(PeersDirectory(serverID), peerID+".toml")
}

func RenderedServerConfigPath(serverID, interfaceName string) string {
	return filepath.Join(ServerDirectory(serverID), "rendered", interfaceName+".conf")
}
