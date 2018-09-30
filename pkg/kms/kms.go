package kms

// Provider interface for key management services
type Provider interface {
	// Store secrets in the KMS at the given path
	Store(path string, s map[string]interface{}) error
}
