package kms

// Provider interface for key management services
type Provider interface {
	StoreAssets(prefix string, assets map[string]interface{}) error
}
