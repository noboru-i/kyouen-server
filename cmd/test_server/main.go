package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"kyouen-server/internal/config"
	"kyouen-server/internal/datastore"
	"kyouen-server/internal/middleware"
	"kyouen-server/internal/statics"
)

type App struct {
	Config           *config.Config
	DatastoreService *datastore.DatastoreService
}

func main() {
	// Set environment variables for testing
	os.Setenv("GOOGLE_CLOUD_PROJECT", "my-android-server")
	os.Setenv("GIN_MODE", "debug")
	// Comment out emulator for production testing
	// os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")
	
	// Load configuration
	cfg := config.Load()
	
	// Initialize Datastore service
	datastoreService, err := datastore.NewDatastoreService(cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize Datastore service: %v", err)
	}
	defer datastoreService.Close()
	
	// Create application instance
	app := &App{
		Config:           cfg,
		DatastoreService: datastoreService,
	}
	
	// Set Gin mode
	gin.SetMode(cfg.Environment)
	
	// Initialize Gin router
	router := setupRouter(app)
	
	// Start server
	log.Printf("Test server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Project ID: %s", cfg.ProjectID)
	log.Println("Available endpoints:")
	log.Println("  GET  /health")
	log.Println("  GET  /v2/statics")
	log.Println("  GET  /v2/stages")
	log.Println("  POST /v2/stages")
	log.Println("  POST /v2/stages/{stageNo}/clear")
	log.Println("  POST /v2/users/login")
	
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(app *App) *gin.Engine {
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
			"status":    "ok",
			"message":   "Cloud Run + Firestore migration test server",
			"version":   "2.0.0-alpha",
			"endpoints": []string{"/health", "/v2/statics", "/v2/stages", "/v2/users/login"},
		})
	})
	
	// Initialize handlers
	staticsHandler := statics.NewHandler(app.DatastoreService)
	
	// API v2 routes
	v2 := router.Group("/v2")
	{
		// Statistics endpoint
		v2.GET("/statics", staticsHandler.GetStatics)
		
		// Note: Stages endpoints are disabled in test server for simplicity
		// They would require Firebase authentication setup
		
		// Users endpoints (test login without Firebase)
		users := v2.Group("/users")
		{
			users.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"screenName": "test_user",
					"token":      "test_token",
					"message":    "Test login - Firebase auth disabled",
				})
			})
		}
	}
	
	return router
}