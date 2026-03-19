package ports

type KeyPair struct {
	PrivateKeyRef string
	PublicKey     string
}

type KeyService interface {
	GenerateKeyPair(scope string) (KeyPair, error)
	GeneratePresharedKeyRef(scope string) (string, error)
}
