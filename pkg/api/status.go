package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yodo-io/cryptogen/pkg/store"
)

// Get task status, will respond with a taskStatusResponse
func (a *api) getStatus(c *gin.Context) {
	jobID := c.Param("jobID")
	s, err := a.store.GetStatus(jobID)
	switch {
	case err == store.ErrNotFound:
		c.JSON(http.StatusNotFound, "job not found")
	case err != nil:
		fatal(c, err)
	default:
		c.JSON(http.StatusOK, s)
	}
}
