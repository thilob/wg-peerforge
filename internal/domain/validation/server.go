package validation

import (
	"regexp"
	"strings"

	"github.com/thilob/wg-peerforge/internal/domain/models"
)

var interfacePattern = regexp.MustCompile(`^[a-zA-Z0-9_=+.-]{1,15}$`)

func ValidateServer(server models.Server) Result {
	errors := []string{}

	if strings.TrimSpace(server.Name) == "" {
		errors = append(errors, "server name is required")
	}
	if !interfacePattern.MatchString(server.InterfaceName) {
		errors = append(errors, "interface name must be 1-15 safe characters")
	}
	if server.ListenPort < 1 || server.ListenPort > 65535 {
		errors = append(errors, "listen port must be between 1 and 65535")
	}
	if !strings.Contains(server.AddressV4, "/") {
		errors = append(errors, "IPv4 tunnel address must use CIDR notation")
	}
	if strings.TrimSpace(server.Endpoint) == "" {
		errors = append(errors, "endpoint is required")
	}
	if strings.TrimSpace(server.PrivateKeyRef) == "" {
		errors = append(errors, "server private key reference is required")
	}
	if strings.TrimSpace(server.PublicKey) == "" {
		errors = append(errors, "server public key is required")
	}

	return Result{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}
