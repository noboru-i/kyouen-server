package stage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"kyouen-server/internal/auth"
)

// StageServiceInterface defines the interface for stage service
type StageServiceInterface interface {
	DeleteAccount(userUID string) error
	GetActivities(limit int) ([]ActivityUser, error)
}

// MockStageService is a simple mock implementation of StageServiceInterface
type MockStageService struct {
	deleteAccountFunc    func(userUID string) error
	deleteAccountCalls  []string
	getActivitiesFunc   func(limit int) ([]ActivityUser, error)
}

func (m *MockStageService) DeleteAccount(userUID string) error {
	m.deleteAccountCalls = append(m.deleteAccountCalls, userUID)
	if m.deleteAccountFunc != nil {
		return m.deleteAccountFunc(userUID)
	}
	return nil
}

func (m *MockStageService) SetDeleteAccountFunc(f func(userUID string) error) {
	m.deleteAccountFunc = f
}

func (m *MockStageService) WasDeleteAccountCalled(userUID string) bool {
	for _, call := range m.deleteAccountCalls {
		if call == userUID {
			return true
		}
	}
	return false
}

func (m *MockStageService) GetActivities(limit int) ([]ActivityUser, error) {
	if m.getActivitiesFunc != nil {
		return m.getActivitiesFunc(limit)
	}
	return []ActivityUser{}, nil
}

func (m *MockStageService) SetGetActivitiesFunc(f func(limit int) ([]ActivityUser, error)) {
	m.getActivitiesFunc = f
}

// TestHandler is a test version of Handler that accepts interface
type TestHandler struct {
	stageService StageServiceInterface
}

func (h *TestHandler) DeleteAccount(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}

func (h *TestHandler) GetActivities(c *gin.Context) {
	activities, err := h.stageService.GetActivities(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]ActivityUserResponse, 0, len(activities))
	for _, a := range activities {
		cs := make([]ActivityStageResponse, len(a.ClearedStages))
		for i, s := range a.ClearedStages {
			cs[i] = ActivityStageResponse{StageNo: s.StageNo, ClearDate: s.ClearDate}
		}
		resp = append(resp, ActivityUserResponse{
			ScreenName:    a.ScreenName,
			Image:         a.Image,
			ClearedStages: cs,
		})
	}

	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, resp)
}

func TestDeleteAccount_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := &MockStageService{}

	// Create test handler
	handler := &TestHandler{
		stageService: mockService,
	}

	// Create router
	router := gin.New()
	router.DELETE("/v2/users/delete-account", func(c *gin.Context) {
		// Set authenticated user in context (simulating middleware)
		c.Set(auth.AuthUIDKey, "test-uid")
		handler.DeleteAccount(c)
	})

	// Create request
	req, _ := http.NewRequest("DELETE", "/v2/users/delete-account", nil)
	resp := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(resp, req)

	// Assert response
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
	
	body := resp.Body.String()
	if !strings.Contains(body, "Account deleted successfully") {
		t.Errorf("Expected response to contain 'Account deleted successfully', got: %s", body)
	}

	// Assert mock was called
	if !mockService.WasDeleteAccountCalled("test-uid") {
		t.Errorf("Expected DeleteAccount to be called with 'test-uid'")
	}
}

func TestGetActivities_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	clearDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	mockService := &MockStageService{}
	mockService.SetGetActivitiesFunc(func(limit int) ([]ActivityUser, error) {
		return []ActivityUser{
			{
				UserID:     "user1",
				ScreenName: "alice",
				Image:      "https://example.com/alice.jpg",
				ClearedStages: []ActivityStage{
					{StageNo: 42, ClearDate: clearDate},
				},
			},
			{
				UserID:     "user2",
				ScreenName: "bob",
				Image:      "https://example.com/bob.jpg",
				ClearedStages: []ActivityStage{
					{StageNo: 10, ClearDate: clearDate},
				},
			},
		}, nil
	})

	handler := &TestHandler{stageService: mockService}
	router := gin.New()
	router.GET("/v2/activities", handler.GetActivities)

	req, _ := http.NewRequest("GET", "/v2/activities", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Header().Get("Cache-Control") != "public, max-age=60" {
		t.Errorf("Expected Cache-Control header, got: %s", resp.Header().Get("Cache-Control"))
	}

	body := resp.Body.String()
	for _, want := range []string{"screen_name", "cleared_stages", "stage_no", "clear_date", "alice", "bob"} {
		if !strings.Contains(body, want) {
			t.Errorf("Expected response to contain %q, got: %s", want, body)
		}
	}
}

func TestGetActivities_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockStageService{}
	mockService.SetGetActivitiesFunc(func(limit int) ([]ActivityUser, error) {
		return nil, errors.New("datastore error")
	})

	handler := &TestHandler{stageService: mockService}
	router := gin.New()
	router.GET("/v2/activities", handler.GetActivities)

	req, _ := http.NewRequest("GET", "/v2/activities", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
	}
}

func TestGetActivities_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockStageService{}

	handler := &TestHandler{stageService: mockService}
	router := gin.New()
	router.GET("/v2/activities", handler.GetActivities)

	req, _ := http.NewRequest("GET", "/v2/activities", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array [], got: %s", body)
	}
}

func TestDeleteAccount_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := &MockStageService{}

	// Create test handler
	handler := &TestHandler{
		stageService: mockService,
	}

	// Create router without authentication
	router := gin.New()
	router.DELETE("/v2/users/delete-account", handler.DeleteAccount)

	// Create request
	req, _ := http.NewRequest("DELETE", "/v2/users/delete-account", nil)
	resp := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(resp, req)

	// Assert response
	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
	
	body := resp.Body.String()
	if !strings.Contains(body, "authentication required") {
		t.Errorf("Expected response to contain 'authentication required', got: %s", body)
	}

	// Assert mock was not called
	if len(mockService.deleteAccountCalls) != 0 {
		t.Errorf("Expected DeleteAccount not to be called, but it was called %d times", len(mockService.deleteAccountCalls))
	}
}

func TestDeleteAccount_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := &MockStageService{}
	testError := errors.New("service error")
	mockService.SetDeleteAccountFunc(func(userUID string) error {
		return testError
	})

	// Create test handler
	handler := &TestHandler{
		stageService: mockService,
	}

	// Create router
	router := gin.New()
	router.DELETE("/v2/users/delete-account", func(c *gin.Context) {
		// Set authenticated user in context (simulating middleware)
		c.Set(auth.AuthUIDKey, "test-uid")
		handler.DeleteAccount(c)
	})

	// Create request
	req, _ := http.NewRequest("DELETE", "/v2/users/delete-account", nil)
	resp := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(resp, req)

	// Assert response
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
	}
	
	body := resp.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("Expected response to contain 'error', got: %s", body)
	}

	// Assert mock was called
	if !mockService.WasDeleteAccountCalled("test-uid") {
		t.Errorf("Expected DeleteAccount to be called with 'test-uid'")
	}
}