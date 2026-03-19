package usecases

import (
	"fmt"
	"strings"

	"github.com/thilob/wg-peerforge/internal/app/ports"
	"github.com/thilob/wg-peerforge/internal/domain/models"
	"github.com/thilob/wg-peerforge/internal/domain/validation"
)

type CreateServerInput struct {
	ID                string
	Name              string
	InterfaceName     string
	ListenPort        int
	Endpoint          string
	AddressV4         string
	AddressV6         string
	DNS               []string
	DefaultAllowedIPs []string
	BackendMode       models.RenderBackend
}

func CreateServer(input CreateServerInput, deps struct {
	ConfigStore ports.ConfigStore
	KeyService  ports.KeyService
}) (models.Server, error) {
	keyPair, err := deps.KeyService.GenerateKeyPair("server:" + input.ID)
	if err != nil {
		return models.Server{}, fmt.Errorf("generate server key pair: %w", err)
	}

	server := models.Server{
		ID:                input.ID,
		Name:              input.Name,
		InterfaceName:     input.InterfaceName,
		ListenPort:        input.ListenPort,
		Endpoint:          input.Endpoint,
		AddressV4:         input.AddressV4,
		AddressV6:         input.AddressV6,
		DNS:               defaultStrings(input.DNS),
		PrivateKeyRef:     keyPair.PrivateKeyRef,
		PublicKey:         keyPair.PublicKey,
		DefaultAllowedIPs: defaultAllowedIPs(input.DefaultAllowedIPs),
		BackendMode:       input.BackendMode,
	}

	result := validation.ValidateServer(server)
	if !result.Valid {
		return models.Server{}, fmt.Errorf(strings.Join(result.Errors, "; "))
	}

	if err := deps.ConfigStore.SaveServer(server); err != nil {
		return models.Server{}, fmt.Errorf("save server: %w", err)
	}

	return server, nil
}

func defaultAllowedIPs(values []string) []string {
	if len(values) == 0 {
		return []string{"0.0.0.0/0"}
	}
	return values
}

func defaultStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return values
}
