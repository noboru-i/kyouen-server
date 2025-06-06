package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kyouen-server/services"
)

func GetStatics(datastoreService *services.DatastoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := datastoreService.GetSummary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"count":        summary.Count,
			"lastUpdatedAt": summary.LastDate,
		})
	}
}