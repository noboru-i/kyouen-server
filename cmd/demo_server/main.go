package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"kyouen-server/internal/config"
	"kyouen-server/internal/middleware"
	"kyouen-server/internal/generated/openapi"
)

func main() {
	// Set environment variables for demo
	os.Setenv("GOOGLE_CLOUD_PROJECT", "demo-project")
	os.Setenv("GIN_MODE", "debug")
	
	// Load configuration
	cfg := config.Load()
	
	// Set Gin mode
	gin.SetMode(cfg.Environment)
	
	// Initialize Gin router
	router := setupRouter(cfg)
	
	// Start server
	log.Printf("Demo server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Project ID: %s", cfg.ProjectID)
	log.Println("Available endpoints:")
	log.Println("  GET  /health")
	log.Println("  GET  /v2/statics")
	log.Println("  GET  /v2/stages")
	log.Println("  POST /v2/stages")
	log.Println("  POST /v2/stages/{stageNo}/clear")
	log.Println("  POST /v2/users/login")
	log.Println()
	log.Println("Demo mode: Returns sample data to demonstrate Cloud Run + Datastore integration")
	
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	
	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	
	// Logging middleware
	router.Use(middleware.Logger())
	
	// Recovery middleware
	router.Use(gin.Recovery())
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"message":     "Cloud Run + Datastore migration demo server",
			"version":     "2.0.0-beta",
			"architecture": "Gin + Datastore + Cloud Run",
			"endpoints":   []string{"/health", "/v2/statics", "/v2/stages", "/v2/users/login"},
		})
	})
	
	// API v2 routes with demo data
	v2 := router.Group("/v2")
	{
		// Statistics endpoint
		v2.GET("/statics", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"count":        42,
				"lastUpdatedAt": time.Now().Add(-30 * time.Minute),
				"note":         "Demo data - integrated with existing Datastore architecture",
			})
		})
		
		// Stages endpoints
		stages := v2.Group("/stages")
		{
			stages.GET("", func(c *gin.Context) {
				now := time.Now()
				c.JSON(http.StatusOK, []openapi.Stage{
					{
						StageNo:    1,
						Size:       5,
						Stage:      "0000010000100001000010000",
						Creator:    "system",
						RegistDate: now.Add(-2 * time.Hour),
					},
					{
						StageNo:    2,
						Size:       5,
						Stage:      "0000110000100001000010000",
						Creator:    "demo_user",
						RegistDate: now.Add(-1 * time.Hour),
					},
				})
			})
			
			stages.POST("", func(c *gin.Context) {
				var param openapi.NewStage
				if err := c.ShouldBindJSON(&param); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(http.StatusCreated, openapi.Stage{
					StageNo:    3,
					Size:       param.Size,
					Stage:      param.Stage,
					Creator:    param.Creator,
					RegistDate: time.Now(),
				})
			})
			
			stages.POST("/:stageNo/clear", func(c *gin.Context) {
				stageNo := c.Param("stageNo")
				
				var param openapi.ClearStage
				if err := c.ShouldBindJSON(&param); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(http.StatusOK, gin.H{
					"stageNo": stageNo,
					"status":  "cleared",
					"message": "Stage cleared successfully (demo mode)",
					"note":    "In production, this validates kyouen and updates user progress",
				})
			})
		}
		
		// Users endpoints
		users := v2.Group("/users")
		{
			users.POST("/login", func(c *gin.Context) {
				var param openapi.LoginParam
				if err := c.ShouldBindJSON(&param); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(http.StatusOK, openapi.LoginResult{
					ScreenName: "demo_user",
					Token:      "demo_firebase_token_here",
				})
			})
		}
	}
	
	return router
}