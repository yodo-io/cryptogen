package crypto

type Peers struct {
	Count int `json:"Count"     yaml:"Count"     binding:"required"`
}

type Users struct {
	Count int `json:"Count"     yaml:"Count"     binding:"required"`
}

type PeerOrg struct {
	Name     string `json:"Name"      yaml:"Name"      binding:"required"`
	Domain   string `json:"Domain"    yaml:"Domain"    binding:"required"`
	Template Peers  `json:"Template"  yaml:"Template"  binding:"required"`
	Users    Users  `json:"Users"     yaml:"Users"     binding:"required"`
}

type GenerateCryptoRequest struct {
	PeerOrgs []PeerOrg `json:"PeerOrgs" yaml:"PeerOrgs"`
}

// Assets is basically a list of secrets, each to be stored under a certain path
type Assets []struct {
	Path    string
	Secrets Secrets
}

// Secrets is a map of key to binary data to be stored in KMS
type Secrets map[string]interface{}

type Provider interface {
	GenerateAssets(taskID string, req GenerateCryptoRequest) (Assets, error)
}
