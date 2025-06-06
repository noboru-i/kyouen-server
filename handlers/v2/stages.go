package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"kyouen-server/config"
	"kyouen-server/db"
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
		
		// Create stage
		newStage := db.KyouenPuzzle{
			Size:    param.Size,
			Stage:   param.Stage,
			Creator: param.Creator,
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
		stage, _, err := datastoreService.GetStageByNo(stageNo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage not found"})
			return
		}
		
		// Verify stage matches
		if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage mismatch"})
			return
		}
		
		// For now, return success - TODO: implement user authentication and stage user creation
		c.JSON(http.StatusOK, gin.H{
			"stageNo": stageNo,
			"status":  "cleared",
			"message": "Stage cleared successfully",
		})
	}
}

func Login(datastoreService *services.DatastoreService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var param openapi.LoginParam
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// TODO: Implement Twitter OAuth and Firebase token generation
		// For now, return a placeholder response
		c.JSON(http.StatusOK, gin.H{
			"screenName": "placeholder_user",
			"token":      "placeholder_token",
			"message":    "Login endpoint implementation pending",
		})
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