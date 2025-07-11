package stage

import (
	"net/http"
	"strconv"
	"time"

	"kyouen-server/internal/auth"
	"kyouen-server/internal/datastore"
	"kyouen-server/internal/generated/openapi"
	"kyouen-server/pkg/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	stageService     *Service
	datastoreService *datastore.DatastoreService
	firebaseService  *datastore.FirebaseService
}

func NewHandler(datastoreService *datastore.DatastoreService, firebaseService *datastore.FirebaseService) *Handler {
	return &Handler{
		stageService:     NewService(datastoreService, firebaseService),
		datastoreService: datastoreService,
		firebaseService:  firebaseService,
	}
}

func (h *Handler) GetStages(c *gin.Context) {
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

	stages, err := h.datastoreService.GetStages(startStageNo, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	stageList := []openapi.Stage{}
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

func (h *Handler) CreateStage(c *gin.Context) {
	// Get authenticated user from context
	authUser, exists := auth.GetAuthenticatedUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var param openapi.NewStage
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	savedStage, err := h.stageService.CreateStage(param, authUser.Name)
	if err != nil {
		switch err {
		case ErrInsufficientStones:
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage must have 5 stones."})
		case ErrNoKyouen:
			c.JSON(http.StatusBadRequest, gin.H{"error": "sent stage don't have kyouen."})
		case ErrStageExists:
			c.JSON(http.StatusConflict, gin.H{"error": "sent stage is already exists."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
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

func (h *Handler) ClearStage(c *gin.Context) {
	// Get authenticated user from context
	authUID, exists := auth.GetAuthenticatedUID(c)
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

	user, err := h.stageService.ClearStage(stageNo, param.Stage, authUID)
	if err != nil {
		switch err {
		case ErrInvalidKyouen:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kyouen"})
		case ErrStageNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage not found"})
		case ErrStageMismatch:
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage mismatch"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stageNo": stageNo,
		"status":  "cleared",
		"message": "Stage cleared successfully",
		"user":    user.ScreenName,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var param openapi.LoginParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Firebase ID token
	ctx := c.Request.Context()
	token, err := h.firebaseService.VerifyIDToken(ctx, param.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Firebase ID token",
		})
		return
	}

	// Get user information from Firebase Auth
	userRecord, err := h.firebaseService.GetUserByUID(ctx, token.UID)
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
	user, err := h.datastoreService.CreateOrUpdateUserFromFirebase(
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
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) SyncStages(c *gin.Context) {
	// Get authenticated user from context
	authUID, exists := auth.GetAuthenticatedUID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var clientClearedStages []openapi.ClearedStage
	if err := c.ShouldBindJSON(&clientClearedStages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serverClearedStages, err := h.stageService.SyncStages(authUID, clientClearedStages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h *Handler) DeleteAccount(c *gin.Context) {
	// Get authenticated user from context
	authUID, exists := auth.GetAuthenticatedUID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	err := h.stageService.DeleteAccount(authUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, openapi.DeleteAccountResult{
		Message: "Account deleted successfully",
	})
}

func (h *Handler) GetRecentStages(c *gin.Context) {
	stages, err := h.datastoreService.GetRecentStages(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	stageList := []openapi.Stage{}
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

type ActivityStage struct {
	StageNo int64     `json:"stageNo"`
	ClearDate time.Time `json:"clearDate"`
}

type ActivityUser struct {
	ScreenName    string          `json:"screenName"`
	Image         string          `json:"image"`
	ClearedStages []ActivityStage `json:"clearedStages"`
}

func (h *Handler) GetActivities(c *gin.Context) {
	stageUsers, err := h.datastoreService.GetRecentActivities(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group activities by user
	userActivities := make(map[string]*ActivityUser)
	for _, stageUser := range stageUsers {
		// Get user information
		user, err := h.datastoreService.GetUserByKey(stageUser.UserKey)
		if err != nil {
			continue // Skip if user not found
		}

		// Get stage information
		stage, err := h.datastoreService.GetStageByKey(stageUser.StageKey)
		if err != nil {
			continue // Skip if stage not found
		}

		// Create or update user activity
		if _, exists := userActivities[user.UserID]; !exists {
			userActivities[user.UserID] = &ActivityUser{
				ScreenName:    user.ScreenName,
				Image:         user.Image,
				ClearedStages: []ActivityStage{},
			}
		}

		userActivities[user.UserID].ClearedStages = append(
			userActivities[user.UserID].ClearedStages,
			ActivityStage{
				StageNo:   stage.StageNo,
				ClearDate: stageUser.ClearDate,
			},
		)
	}

	// Convert to response format
	var activities []ActivityUser
	for _, activity := range userActivities {
		activities = append(activities, *activity)
	}

	c.JSON(http.StatusOK, activities)
}

// Helper function for validation
func isKyouen(kyouenStage *models.KyouenStage) bool {
	return kyouenStage.IsKyouenByWhite() != nil
}
