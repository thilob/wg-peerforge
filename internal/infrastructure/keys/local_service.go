package keys

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/thilob/wg-peerforge/internal/app/ports"
)

type LocalService struct{}

func NewLocalService() LocalService {
	return LocalService{}
}

func (LocalService) GenerateKeyPair(scope string) (ports.KeyPair, error) {
	privateRefToken, err := randomToken()
	if err != nil {
		return ports.KeyPair{}, fmt.Errorf("generate private key token for %s: %w", scope, err)
	}

	publicKeyToken, err := randomToken()
	if err != nil {
		return ports.KeyPair{}, fmt.Errorf("generate public key token for %s: %w", scope, err)
	}

	return ports.KeyPair{
		PrivateKeyRef: fmt.Sprintf("local:%s:%s", sanitizeScope(scope), privateRefToken),
		PublicKey:     publicKeyToken,
	}, nil
}

func (LocalService) GeneratePresharedKeyRef(scope string) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate preshared key token for %s: %w", scope, err)
	}

	return fmt.Sprintf("local:%s:psk:%s", sanitizeScope(scope), token), nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func sanitizeScope(scope string) string {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return "default"
	}

	scope = strings.ReplaceAll(scope, ":", "_")
	scope = strings.ReplaceAll(scope, "/", "_")
	return scope
}
