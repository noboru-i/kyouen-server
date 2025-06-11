package statics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kyouen-server/internal/datastore"
)

type Handler struct {
	datastoreService *datastore.DatastoreService
}

func NewHandler(datastoreService *datastore.DatastoreService) *Handler {
	return &Handler{
		datastoreService: datastoreService,
	}
}

func (h *Handler) GetStatics(c *gin.Context) {
	summary, err := h.datastoreService.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"count":        summary.Count,
		"lastUpdatedAt": summary.LastDate,
	})
}