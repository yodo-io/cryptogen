package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis"
	vault "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
)

type CryptogenPeers struct {
	Count int `json:"Count"     yaml:"Count"     binding:"required"`
}

type CryptogenUsers struct {
	Count int `json:"Count"     yaml:"Count"     binding:"required"`
}

type CryptogenOrg struct {
	Name     string         `json:"Name"      yaml:"Name"      binding:"required"`
	Domain   string         `json:"Domain"    yaml:"Domain"    binding:"required"`
	Template CryptogenPeers `json:"Template"  yaml:"Template"  binding:"required"`
	Users    CryptogenUsers `json:"Users"     yaml:"Users"     binding:"required"`
}

type GenerateCryptoRequest struct {
	PeerOrgs []CryptogenOrg `json:"PeerOrgs" yaml:"PeerOrgs"`
}

type GenerateCryptoResponse struct {
	TaskID string `json:"TaskID"`
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Env var %s must be set", key)
	}
	return v
}

type Cryptogen struct {
	vault  *vault.Logical
	redis  *redis.Client
	tmpDir string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	addr := mustGetEnv("ADDRESS")
	tmpDir := mustGetEnv("TMP_DIR")
	vaultTokenPath := mustGetEnv("VAULT_TOKEN_PATH")
	redisAddr := mustGetEnv("REDIS_ADDR")

	// It's VAULT_ADDR, not VAULT_ADDRESS
	if _, isSet := os.LookupEnv("VAULT_ADDR"); !isSet {
		log.Print("Warning: VAULT_ADDR not set, make sure it is spelled correctly. Will default to https://127.0.0.1:8200")
	}

	token, err := ioutil.ReadFile(vaultTokenPath)
	if err != nil {
		log.Fatalf("Failed to read vault token: %v", err)
	}

	server(addr, tmpDir, string(token), redisAddr)
}

func server(addr, tmpDir, vaultToken, redisAddr string) {
	// vault
	cnf := vault.DefaultConfig()
	vc, err := vault.NewClient(cnf)
	if err != nil {
		log.Fatalf("Failed to init vault client: %v", err)
	}
	log.Printf("Using vault token %s", vaultToken)
	vc.SetToken(vaultToken)

	// redis
	rc := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rc.Ping().Result()
	if err != nil {
		log.Fatalf("Failed connecting to redis %v", err)
	}
	log.Printf("Redis pong %s", pong)

	s := &Cryptogen{
		vault:  vc.Logical(),
		redis:  rc,
		tmpDir: tmpDir,
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "i'm fine")
	})

	r.GET("/task/:taskID", func(c *gin.Context) {
		taskID := c.Param("taskID")
		status, err := s.GetTaskStatus(taskID)
		if err == ErrTaskNotFound {
			c.JSON(http.StatusNotFound, "task not found")
			return
		}
		if err != nil {
			fatal(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
	})

	r.POST("/crypto-assets", func(c *gin.Context) {
		var req GenerateCryptoRequest
		if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
			log.Print(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		taskID, err := s.GenerateCrypto(req)
		if err != nil {
			fatal(c, err)
			return
		}
		c.JSON(http.StatusAccepted, GenerateCryptoResponse{TaskID: taskID})
	})

	r.Run(addr)
}

func (cg *Cryptogen) doGenerateAssets(taskID string, req GenerateCryptoRequest) (<-chan bool, <-chan error) {
	errch := make(chan error)
	donech := make(chan bool)

	go func() {
		res, err := cg.generateCryptoAssets(taskID, req)
		if err != nil {
			errch <- err
			return
		}
		if err := cg.storeInVault(taskID, res); err != nil {
			errch <- err
			return
		}
		donech <- true
	}()

	return donech, errch

}

// basically a list of file paths
type cryptoAssets map[string]interface{}

func (cg *Cryptogen) generateCryptoAssets(taskID string, req GenerateCryptoRequest) (cryptoAssets, error) {
	prefix := path.Join(cg.tmpDir, taskID)
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
		"/usr/local/bin/cryptogen",
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

	assets := cryptoAssets{
		"foo": "bar",
	}

	log.Printf("Crypto assets generated and written to %s", prefix)
	return assets, nil
}

// store given assets in vault
func (cg *Cryptogen) storeInVault(taskID string, ca cryptoAssets) error {
	vaultPath := "/secret/cryptogen/" + taskID
	if _, err := cg.vault.Write(vaultPath, ca); err != nil {
		return fmt.Errorf("Error writing to vault: %v", err)
	}
	log.Printf("Secrets written to vault path %s", vaultPath)
	return nil
}

func taskID() string {
	return fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(100000))
}

func taskKey(taskID string) string {
	return fmt.Sprintf("cryptogen.task.%s", taskID)
}

const (
	TaskStatusNew        string = "new"
	TaskStatusProcessing string = "processing"
	TaskStatusComplete   string = "complete"
)

// GenerateCrypto generates crypto assets for given request. will process request in
// background, error will only be returned if request is invalid. errors during actual task
// processing will be logged to stderr and task status database will be updated.
// todo: update task status in some kind of backend database.
func (cg *Cryptogen) GenerateCrypto(req GenerateCryptoRequest) (string, error) {
	taskID := taskID()
	key := taskKey(taskID)

	if err := cg.redis.Set(key, TaskStatusNew, 0).Err(); err != nil {
		log.Printf("Error updating task status %v", err)
		return "", err
	}

	donech, errch := cg.doGenerateAssets(taskID, req)

	if err := cg.redis.Set(key, TaskStatusProcessing, 0).Err(); err != nil {
		log.Printf("Error updating task status %v", err)
		return "", err
	}

	go func() {
		select {
		case err := <-errch:
			log.Printf("Error in task %s: %v", taskID, err)
		case <-donech:
			log.Printf("Done task %s", taskID)
			key := taskKey(taskID)
			if err := cg.redis.Set(key, TaskStatusComplete, 0).Err(); err != nil {
				log.Printf("Error updating task %s: %v", taskID, err)
			}
		}
	}()

	return taskID, nil
}

var ErrTaskNotFound = errors.New("TaskNotFound")

// GetTaskStatus returns the current status for a given task
func (cg *Cryptogen) GetTaskStatus(taskID string) (string, error) {
	key := taskKey(taskID)
	val, err := cg.redis.Get(key).Result()
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", ErrTaskNotFound
	}
	return val, nil
}

// Abort gin.Context with http.StatusInternalServerError respond and error message as JSON string
func fatal(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
}
