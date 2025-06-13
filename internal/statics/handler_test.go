package statics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"kyouen-server/internal/datastore"
)

func TestGetStatics(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test Datastore service
	datastoreService, err := datastore.NewDatastoreService("test-project-id")
	if err != nil {
		t.Skipf("Skipping test - Datastore not available: %v", err)
		return
	}
	defer datastoreService.Close()
	
	// Create handler
	handler := NewHandler(datastoreService)
	
	// Create a Gin router
	router := gin.New()
	router.GET("/v2/statics", handler.GetStatics)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/v2/statics", nil)
	resp := httptest.NewRecorder()
	
	// Perform the request
	router.ServeHTTP(resp, req)
	
	// Assert the response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "count")
}