package stage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"kyouen-server/internal/auth"
)

// StageServiceInterface defines the interface for stage service
type StageServiceInterface interface {
	DeleteAccount(userUID string) error
}

// MockStageService is a simple mock implementation of StageServiceInterface
type MockStageService struct {
	deleteAccountFunc func(userUID string) error
	deleteAccountCalls []string
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