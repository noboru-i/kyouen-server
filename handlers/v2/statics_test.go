package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"kyouen-server/services"
)

func TestGetStatics(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test Datastore service
	datastoreService, err := services.NewDatastoreService("test-project-id")
	if err != nil {
		t.Skipf("Skipping test - Datastore not available: %v", err)
		return
	}
	defer datastoreService.Close()
	
	// Create a Gin router
	router := gin.New()
	router.GET("/v2/statics", GetStatics(datastoreService))
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/v2/statics", nil)
	resp := httptest.NewRecorder()
	
	// Perform the request
	router.ServeHTTP(resp, req)
	
	// Assert the response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "count")
}