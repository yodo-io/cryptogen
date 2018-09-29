package kms

import (
	"fmt"
	"io/ioutil"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/yodo-io/cryptogen/pkg/crypto"
)

// Vault implementation of kms.Provider interface
type Vault struct {
	client *vault.Logical
}

// VaultConf is the config struct for kms.NewVault
type VaultConf struct {
	TokenPath string
	Token     string
}

// NewVault creates a new instance of kms.Vault
// Returns an error if anything goes wrong during setup
func NewVault(c VaultConf) (*Vault, error) {
	token, err := getToken(c)
	if err != nil {
		return nil, err
	}

	cnf := vault.DefaultConfig()
	vc, err := vault.NewClient(cnf)
	if err != nil {
		return nil, err
	}
	vc.SetToken(token)

	return &Vault{
		client: vc.Logical(),
	}, nil
}

func getToken(c VaultConf) (string, error) {
	switch {
	case c.Token != "":
		return c.Token, nil
	case c.TokenPath != "":
		return readTokenFromFile(c.TokenPath)
	default:
		return "", fmt.Errorf("No token source in config")
	}
}

func readTokenFromFile(tokenPath string) (string, error) {
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("Failed to read vault token from %s: %v", tokenPath, err)
	}
	return string(token), nil
}

// StoreAssets stores the given assets in vault at a given path prefix
func (v *Vault) StoreAssets(p string, s crypto.Secrets) error {
	if _, err := v.client.Write(p, s); err != nil {
		return fmt.Errorf("Error writing to vault at path %s: %v", p, err)
	}
	log.Printf("Assets written to vault at path %s", p)
	return nil
}
