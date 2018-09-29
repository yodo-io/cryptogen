package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yodo-io/cryptogen/pkg/crypto/worker"
	"github.com/yodo-io/cryptogen/pkg/store"
)

// Config is the provider configuration to be used by the API
type Config struct {
	Worker *worker.Worker
	Store  store.Provider
}

// api implementation, not to be exported - we only register the
// route handlers it provides with a Gin router
type api struct {
	worker *worker.Worker
	store  store.Provider
}

// New creates a new API instance and gin.Engine
func New(c *Config) *gin.Engine {
	r := gin.Default()
	a := &api{
		worker: c.Worker,
		store:  c.Store,
	}

	r.POST("/crypto-assets", a.genAssets)
	r.GET("/status/:jobID", a.getStatus)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Status": "OK"})
	})

	return r
}

// Abort gin.Context with http.StatusInternalServerError respond and error message as JSON string
func fatal(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
}
