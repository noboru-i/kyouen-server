package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kyouen-server/config"
	"kyouen-server/services"
)

func GetStages(firestoreService *services.FirestoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual stages retrieval from Firestore
		// For now, return a placeholder response
		c.JSON(http.StatusOK, []gin.H{
			{
				"stageNo":    1,
				"size":       5,
				"stage":      "0000010000100001000010000",
				"creator":    "system",
				"registDate": "2024-01-01T00:00:00Z",
			},
		})
	}
}

func CreateStage(firestoreService *services.FirestoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual stage creation in Firestore
		// For now, return a placeholder response
		c.JSON(http.StatusCreated, gin.H{
			"stageNo":    2,
			"size":       5,
			"stage":      "0000010000100001000010000",
			"creator":    "user",
			"registDate": "2024-01-01T00:00:00Z",
			"message":    "Stage creation endpoint migrated to Firestore (placeholder)",
		})
	}
}

func ClearStage(firestoreService *services.FirestoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stageNo := c.Param("stageNo")
		
		// TODO: Implement actual stage clearing logic with Firestore
		// For now, return a placeholder response
		c.JSON(http.StatusOK, gin.H{
			"stageNo": stageNo,
			"status":  "cleared",
			"message": "Stage clear endpoint migrated to Firestore (placeholder)",
		})
	}
}

func Login(firestoreService *services.FirestoreService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual login logic with Twitter OAuth and Firebase
		// For now, return a placeholder response
		c.JSON(http.StatusOK, gin.H{
			"screenName": "placeholder_user",
			"token":      "placeholder_token",
			"message":    "Login endpoint migrated to Firestore (placeholder)",
		})
	}
}