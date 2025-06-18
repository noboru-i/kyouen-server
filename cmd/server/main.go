package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"kyouen-server/internal/auth"
	"kyouen-server/internal/config"
	"kyouen-server/internal/datastore"
	"kyouen-server/internal/middleware"
	"kyouen-server/internal/stage"
	"kyouen-server/internal/statics"
)

type App struct {
	Config           *config.Config
	DatastoreService *datastore.DatastoreService
	FirebaseService  *datastore.FirebaseService
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Datastore service
	datastoreService, err := datastore.NewDatastoreService(cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize Datastore service: %v", err)
	}
	defer datastoreService.Close()

	// Initialize Firebase service
	firebaseService, err := datastore.NewFirebaseService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase service: %v", err)
	}

	// Create application instance
	app := &App{
		Config:           cfg,
		DatastoreService: datastoreService,
		FirebaseService:  firebaseService,
	}

	// Set Gin mode
	gin.SetMode(cfg.Environment)

	// Initialize Gin router
	router := setupRouter(app)

	// Start server
	log.Printf("Cloud Run Kyouen Server starting on port %s", cfg.Port)
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
			"status":      "ok",
			"version":     "2.0.0-cloudrun",
			"service":     "kyouen-server",
			"platform":    "Cloud Run + Datastore mode Firestore",
			"description": "共円パズルゲーム API Server",
		})
	})

	// Initialize handlers
	stageHandler := stage.NewHandler(app.DatastoreService, app.FirebaseService)
	staticsHandler := statics.NewHandler(app.DatastoreService)

	// API v2 routes
	v2 := router.Group("/v2")
	{
		// Statistics endpoint
		v2.GET("/statics", staticsHandler.GetStatics)

		// Stages endpoints
		stages := v2.Group("/stages")
		{
			stages.GET("", stageHandler.GetStages)
			// Protected endpoints requiring authentication
			stages.POST("", auth.FirebaseAuth(app.FirebaseService), stageHandler.CreateStage)
			stages.POST("/:stageNo/clear", auth.FirebaseAuth(app.FirebaseService), stageHandler.ClearStage)
			stages.POST("/sync", auth.FirebaseAuth(app.FirebaseService), stageHandler.SyncStages)
		}

		// Users endpoints
		users := v2.Group("/users")
		{
			users.POST("/login", stageHandler.Login)
		}
	}

	return router
}
