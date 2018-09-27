package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"

	"gopkg.in/yaml.v2"
)

var addr string

type cryptoOrg struct {
	Name     string `json:"Name"    yaml:"Name"      binding:"required"`
	Domain   string `json:"Domain"  yaml:"Domain"    binding:"required"`
	Template struct {
		Count int `json:"Count"     yaml:"Count"     binding:"required"`
	} `            json:"Template"  yaml:"Template"  binding:"required"`
	Users struct {
		Count int `json:"Count"     yaml:"Count"     binding:"required"`
	} `            json:"Users"     yaml:"Users"     binding:"required"`
}

type generateCryptoRequest struct {
	PeerOrgs []cryptoOrg `json:"PeerOrgs" yaml:"PeerOrgs"`
}

type generateCryptoResponse struct {
	TaskID string `json:"TaskID"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&addr, "addr", ":3000", "Address to bind to")
	flag.Parse()

	server(addr)
}

func server(addr string) {
	r := gin.Default()
	r.POST("/crypto-assets", createAssets)
	r.Run(addr)
}

func createAssets(c *gin.Context) {
	var req generateCryptoRequest
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Print(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	taskID, err := generateCryptoAssets(req)
	if err != nil {
		fatal(c, err)
	}

	// response
	c.JSON(http.StatusAccepted, generateCryptoResponse{TaskID: taskID})
}

func generateCryptoAssets(req generateCryptoRequest) (string, error) {

	taskID := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(100000))

	errch := make(chan error)
	donech := make(chan bool)

	// todo: should update tasks status somewhere
	go func() {
		select {
		case err := <-errch:
			log.Printf("Error in task %s: %v", taskID, err)
		case <-donech:
			log.Printf("Done task %s", taskID)
		}
	}()

	// process crypto gen
	go func(req generateCryptoRequest, taskID string) {

		prefix := fmt.Sprintf("./tmp/_cryptogen/" + taskID)
		cryptoConfPath := prefix + "/crypto-config.yaml"
		cryptoOutPath := prefix + "/crypto-config"

		if err := os.MkdirAll(prefix, 0700); err != nil {
			errch <- err
		}

		// generate crypto-config.yaml
		y, err := yaml.Marshal(req)
		if err != nil {
			errch <- err
		}
		if err := ioutil.WriteFile(cryptoConfPath, y, 0600); err != nil {
			errch <- err
		}

		// generate crypto assets
		cmd := exec.Command(
			"./tools/cryptogen",
			"gene",
			"--config="+cryptoConfPath,
			"--output="+cryptoOutPath,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			errch <- err
		}
		if err := cmd.Wait(); err != nil {
			errch <- err
		}

		donech <- true
	}(req, taskID)

	return taskID, nil
}

func fatal(c *gin.Context, err error) {
	c.AbortWithError(http.StatusInternalServerError, err)
}
