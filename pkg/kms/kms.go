package kms

import "github.com/yodo-io/cryptogen/pkg/crypto"

// Provider interface for key management services
type Provider interface {
	StoreAssets(p string, s crypto.Secrets) error
}
