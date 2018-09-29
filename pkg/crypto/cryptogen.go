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

type Provider interface {
	GenerateAssets(taskID string, req GenerateCryptoRequest) (map[string]interface{}, error)
}
