package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yodo-io/cryptogen/pkg/store"
)

type taskStatusResponse struct {
	Status string
	JobID  string
}

// Get task status, will respond with a taskStatusResponse
func (a *api) getStatus(c *gin.Context) {
	jobID := c.Param("jobID")
	status, err := a.store.GetStatus(jobID)
	switch {
	case err == store.ErrNotFound:
		c.JSON(http.StatusNotFound, "job not found")
	case err != nil:
		fatal(c, err)
	default:
		c.JSON(http.StatusOK, taskStatusResponse{
			Status: status,
			JobID:  jobID,
		})
	}
}
