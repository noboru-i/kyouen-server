package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"kyouen-server/db"
	"kyouen-server/middleware"
	"kyouen-server/models"
	"kyouen-server/openapi"
	"kyouen-server/services"
)

func GetStages(datastoreService *services.DatastoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		startStageNo, err := strconv.Atoi(c.DefaultQuery("start_stage_no", "0"))
		if err != nil {
			startStageNo = 0
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
		
		stages, err := datastoreService.GetStages(startStageNo, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Convert to response format
		var stageList []openapi.Stage
		for _, stage := range stages {
			stageList = append(stageList, openapi.Stage{
				StageNo:    stage.StageNo,
				Size:       stage.Size,
				Stage:      stage.Stage,
				Creator:    stage.Creator,
				RegistDate: stage.RegistDate,
			})
		}
		
		c.JSON(http.StatusOK, stageList)
	}
}

func CreateStage(datastoreService *services.DatastoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user from context
		authUser, exists := middleware.GetAuthenticatedUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		
		var param openapi.NewStage
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validate stage using existing business logic
		stage := *models.NewKyouenStage(int(param.Size), param.Stage)
		
		// Check stone count
		if stage.StoneCount() <= 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage must have 5 stones."})
			return
		}
		
		// Check if stage has kyouen
		kyouenData := stage.HasKyouen()
		if kyouenData == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sent stage don't have kyouen."})
			return
		}
		
		// Check if stage already exists (including rotations and reflections)
		exists, err := hasRegisteredStageAll(datastoreService, stage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "sent stage is already exists."})
			return
		}
		
		// Create stage with authenticated user as creator
		newStage := db.KyouenPuzzle{
			Size:    param.Size,
			Stage:   param.Stage,
			Creator: authUser.Name, // Use authenticated user's name as creator
		}
		
		savedStage, err := datastoreService.CreateStage(newStage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Return response
		c.JSON(http.StatusCreated, openapi.Stage{
			StageNo:    savedStage.StageNo,
			Size:       savedStage.Size,
			Stage:      savedStage.Stage,
			Creator:    savedStage.Creator,
			RegistDate: savedStage.RegistDate,
		})
	}
}

func ClearStage(datastoreService *services.DatastoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user from context
		authUID, exists := middleware.GetAuthenticatedUID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		
		stageNo, err := strconv.Atoi(c.Param("stageNo"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stage number"})
			return
		}
		
		var param openapi.ClearStage
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validate clear stage using existing business logic
		size := int(math.Sqrt(float64(len(param.Stage))))
		paramKyouenStage := models.NewKyouenStage(size, param.Stage)
		
		if !isKyouen(paramKyouenStage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kyouen"})
			return
		}
		
		// Get stage from database
		stage, stageKeys, err := datastoreService.GetStageByNo(stageNo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage not found"})
			return
		}
		
		// Verify stage matches
		if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage mismatch"})
			return
		}
		
		// Get user from database
		user, userKey, err := datastoreService.GetUserByID(authUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		
		// Create stage user relation to record the clear
		err = datastoreService.CreateStageUser(stageKeys[0], userKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record stage clear"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"stageNo": stageNo,
			"status":  "cleared",
			"message": "Stage cleared successfully",
			"user":    user.ScreenName,
		})
	}
}

func Login(datastoreService *services.DatastoreService, firebaseService *services.FirebaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var param openapi.LoginParam
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Verify Firebase ID token
		ctx := c.Request.Context()
		token, err := firebaseService.VerifyIDToken(ctx, param.Token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Firebase ID token",
			})
			return
		}
		
		// Get user information from Firebase Auth
		userRecord, err := firebaseService.GetUserByUID(ctx, token.UID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user information",
			})
			return
		}
		
		// Extract Twitter information
		screenName := userRecord.DisplayName
		image := userRecord.PhotoURL
		twitterUID := ""
		
		// Extract Twitter UID from custom claims
		if claims, ok := token.Claims["firebase"].(map[string]interface{}); ok {
			if identities, ok := claims["identities"].(map[string]interface{}); ok {
				if twitterIds, ok := identities["twitter.com"].([]interface{}); ok && len(twitterIds) > 0 {
					if twitterID, ok := twitterIds[0].(string); ok {
						twitterUID = twitterID
					}
				}
			}
		}
		
		// Fallback: try to get screen name from Twitter provider data
		if screenName == "" {
			for _, provider := range userRecord.ProviderUserInfo {
				if provider.ProviderID == "twitter.com" {
					screenName = provider.DisplayName
					if image == "" {
						image = provider.PhotoURL
					}
					if twitterUID == "" {
						twitterUID = provider.UID
					}
					break
				}
			}
		}
		
		// Create or update user in Datastore
		user, err := datastoreService.CreateOrUpdateUserFromFirebase(
			token.UID,
			screenName,
			image,
			twitterUID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create or update user",
			})
			return
		}
		
		// Return successful login response
		response := openapi.LoginResult{
			ScreenName: user.ScreenName,
			Token:      param.Token, // Return the same Firebase ID token
		}
		
		c.JSON(http.StatusOK, response)
	}
}

func SyncStages(datastoreService *services.DatastoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user from context
		authUID, exists := middleware.GetAuthenticatedUID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		
		var clientClearedStages []openapi.ClearedStage
		if err := c.ShouldBindJSON(&clientClearedStages); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Get user from database
		_, userKey, err := datastoreService.GetUserByID(authUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		
		// For each client cleared stage, create stage user relation if not exists
		for _, clearedStage := range clientClearedStages {
			// Get stage by stage number
			_, stageKeys, err := datastoreService.GetStageByNo(int(clearedStage.StageNo))
			if err != nil {
				// Skip stages that don't exist
				continue
			}
			
			// Check if stage user relation already exists
			exists, err := datastoreService.HasStageUser(stageKeys[0], userKey)
			if err != nil {
				continue
			}
			
			if !exists {
				// Create stage user relation
				err = datastoreService.CreateStageUser(stageKeys[0], userKey)
				if err != nil {
					continue
				}
			}
		}
		
		// Get all cleared stages for this user from server
		serverClearedStages, err := datastoreService.GetClearedStagesByUser(userKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cleared stages"})
			return
		}
		
		// Convert to response format
		var response []openapi.ClearedStage
		for _, stageUser := range serverClearedStages {
			response = append(response, openapi.ClearedStage{
				StageNo:   stageUser.StageKey.ID,
				ClearDate: stageUser.ClearDate,
			})
		}
		
		c.JSON(http.StatusOK, response)
	}
}

// Helper functions from original stages_handler.go
func hasRegisteredStageAll(datastoreService *services.DatastoreService, stage models.KyouenStage) (bool, error) {
	for i := 0; i < 4; i++ {
		mirror := models.NewMirroredKyouenStage(stage)
		exists, err := datastoreService.CheckStageExists(mirror.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
		
		stage = *models.NewRotatedKyouenStage(stage)
		exists, err = datastoreService.CheckStageExists(stage.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func isKyouen(kyouenStage *models.KyouenStage) bool {
	return kyouenStage.IsKyouenByWhite() != nil
}