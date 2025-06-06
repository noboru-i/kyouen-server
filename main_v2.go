package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"kyouen-server/config"
	handlers "kyouen-server/handlers/v2"
	"kyouen-server/middleware"
	"kyouen-server/services"
)

type App struct {
	Config           *config.Config
	FirestoreService *services.FirestoreService
}

func mainV2() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize Firestore service
	firestoreService, err := services.NewFirestoreService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore service: %v", err)
	}
	defer firestoreService.Close()
	
	// Create application instance
	app := &App{
		Config:           cfg,
		FirestoreService: firestoreService,
	}
	
	// Set Gin mode
	gin.SetMode(cfg.Environment)
	
	// Initialize Gin router
	router := setupRouter(app)
	
	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Project ID: %s", cfg.ProjectID)
	
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(app *App) *gin.Engine {
	router := gin.Default()
	
	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, specify allowed origins
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
			"timestamp": "2024-01-01T00:00:00Z", // TODO: Use current time
			"version":   "2.0.0",
		})
	})
	
	// API v2 routes
	v2 := router.Group("/v2")
	{
		// Statistics endpoint (simplest to start with)
		v2.GET("/statics", handlers.GetStatics(app.FirestoreService))
		
		// Stages endpoints
		stages := v2.Group("/stages")
		{
			stages.GET("", handlers.GetStages(app.FirestoreService))
			stages.POST("", handlers.CreateStage(app.FirestoreService))
			stages.POST("/:stageNo/clear", handlers.ClearStage(app.FirestoreService))
		}
		
		// Users endpoints
		users := v2.Group("/users")
		{
			users.POST("/login", handlers.Login(app.FirestoreService, app.Config))
		}
	}
	
	return router
}