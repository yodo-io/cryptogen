package crypto

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	TmpDir        string // tmp dir to store assets in
	CryptogenPath string // path to cryptogen binary
}

type Cryptogen struct {
	Config
}

func New(c Config) (*Cryptogen, error) {
	return &Cryptogen{Config: c}, nil
}

func (c *Cryptogen) GenerateAssets(taskID string, req GenerateCryptoRequest) (map[string]interface{}, error) {
	prefix := path.Join(c.Config.TmpDir, taskID)
	cryptoConfPath := prefix + "/crypto-config.yaml"
	cryptoOutPath := prefix + "/crypto-config"

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

	assets := map[string]interface{}{
		"foo": "bar",
	}

	log.Printf("Crypto assets generated and written to %s", prefix)
	return assets, nil
}
