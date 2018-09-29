package crypto

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	yaml "gopkg.in/yaml.v2"
)

// Config for Cryptogen service
type Config struct {
	TmpDir        string // tmp dir to store assets in
	CryptogenPath string // path to cryptogen binary
}

// Cryptogen service
type Cryptogen struct {
	Config
}

// New initialises a new Cryptogen
func New(c Config) (*Cryptogen, error) {
	return &Cryptogen{Config: c}, nil
}

// GenerateAssets generates crypto assets and returns them for further processing
func (c *Cryptogen) GenerateAssets(taskID string, req GenerateCryptoRequest) (Assets, error) {
	prefix := path.Join(c.Config.TmpDir, taskID)
	cryptoConfPath := prefix + "/crypto-config.yaml"
	cryptoOutPath := prefix + "/crypto-config/"

	if err := os.MkdirAll(prefix, 0700); err != nil {
		return nil, err
	}

	// generate crypto-config.yaml
	y, err := yaml.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(cryptoConfPath, y, 0600); err != nil {
		return nil, err
	}

	// generate crypto assets
	cmd := exec.Command(
		c.Config.CryptogenPath,
		"generate",
		"--config="+cryptoConfPath,
		"--output="+cryptoOutPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// push any errors to error channel
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	// read assets
	assets, err := readAssets(cryptoOutPath, req)
	if err != nil {
		return nil, err
	}

	log.Printf("Crypto assets generated and written to %s", prefix)
	return assets, nil
}

// Generate paths to generated crypto assets based on request
func genPaths(req GenerateCryptoRequest) []string {
	paths := []string{}

	for o := 0; o < len(req.PeerOrgs); o++ {
		dn := req.PeerOrgs[o].Domain
		paths = append(paths, fmt.Sprintf("%s/ca", dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/tls", dn, dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/msp/signcerts", dn, dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/msp/tlscacerts", dn, dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/msp/admincerts", dn, dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/msp/keystore", dn, dn))
		paths = append(paths, fmt.Sprintf("%s/users/Admin@%s/msp/cacerts", dn, dn))

		for p := 0; p < req.PeerOrgs[o].Template.Count; p++ {
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/tls", dn, o, dn))
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/msp/signcerts", dn, o, dn))
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/msp/tlscacerts", dn, o, dn))
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/msp/admincerts", dn, o, dn))
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/msp/keystore", dn, o, dn))
			paths = append(paths, fmt.Sprintf("%s/peers/peer%d.%s/msp/cacerts", dn, o, dn))
		}

		for u := 1; u <= req.PeerOrgs[o].Users.Count; u++ {
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/tls", dn, u, dn))
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/msp/signcerts", dn, u, dn))
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/msp/tlscacerts", dn, u, dn))
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/msp/admincerts", dn, u, dn))
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/msp/keystore", dn, u, dn))
			paths = append(paths, fmt.Sprintf("%s/users/User%d@%s/msp/cacerts", dn, u, dn))
		}
	}
	return paths
}

// Note: a lot of the crypto assets generated are duplicated. For simplicities sake, we still
// just read all of this to resemble the structure created by cryptogen as closely as possible
func readAssets(prefix string, req GenerateCryptoRequest) (Assets, error) {
	// var out Assets

	peersPrefix := path.Join(prefix, "peerOrganizations")
	paths := genPaths(req)
	assets := make(Assets, len(paths))

	for i, p := range paths {
		secrets := Secrets{}

		dir := path.Join(peersPrefix, p)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			n := file.Name()
			fp := path.Join(dir, n)

			b, err := ioutil.ReadFile(fp)
			if err != nil {
				return nil, err
			}
			log.Printf("read %d bytes from %s", len(b), fp)
			secrets[n] = b
		}
		assets[i].Secrets = secrets
		assets[i].Path = path.Join("peerOrganizations", p)
	}

	return assets, nil
}
