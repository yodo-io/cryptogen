package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/yodo-io/cryptogen/pkg/crypto"
)

type generateCryptoResponse struct {
	JobID string
}

func (a *api) genAssets(c *gin.Context) {
	var req crypto.GenerateCryptoRequest
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Print(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	job := a.worker.NewJob(req)
	c.JSON(http.StatusAccepted, generateCryptoResponse{JobID: job.ID})
}
