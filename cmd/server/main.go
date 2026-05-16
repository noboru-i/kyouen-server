package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"kyouen-server/internal/auth"
	"kyouen-server/internal/config"
	"kyouen-server/internal/datastore"
	"kyouen-server/internal/middleware"
	"kyouen-server/internal/stage"
	"kyouen-server/internal/statics"
	"kyouen-server/internal/tracing"
)

type App struct {
	Config           *config.Config
	DatastoreService *datastore.DatastoreService
	FirebaseService  *datastore.FirebaseService
}

func main() {
	ctx := context.Background()

	cfg := config.Load()

	shutdown := tracing.Init(ctx, cfg.ProjectID)
	defer shutdown()

	datastoreService, err := datastore.NewDatastoreService(cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize Datastore service: %v", err)
	}
	defer datastoreService.Close()

	firebaseService, err := datastore.NewFirebaseService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase service: %v", err)
	}

	app := &App{
		Config:           cfg,
		DatastoreService: datastoreService,
		FirebaseService:  firebaseService,
	}

	gin.SetMode(cfg.Environment)

	router := setupRouter(app)

	log.Printf("Cloud Run Kyouen Server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Project ID: %s", cfg.ProjectID)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(app *App) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.Use(otelgin.Middleware("kyouen-server"))
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"version":     "2.0.0-cloudrun",
			"service":     "kyouen-server",
			"platform":    "Cloud Run + Datastore mode Firestore",
			"description": "共円パズルゲーム API Server",
		})
	})

	router.StaticFile("/docs/specs/index.yaml", "./docs/specs/index.yaml")
	router.StaticFile("/static/swagger-ui.html", "./static-files/swagger-ui.html")

	stageHandler := stage.NewHandler(app.DatastoreService, app.FirebaseService)
	staticsHandler := statics.NewHandler(app.DatastoreService)

	v2 := router.Group("/v2")
	{
		v2.GET("/statics", staticsHandler.GetStatics)

		v2.GET("/recent_stages", stageHandler.GetRecentStages)
		v2.GET("/activities", stageHandler.GetActivities)

		stages := v2.Group("/stages")
		{
			stages.GET("", auth.OptionalFirebaseAuth(app.FirebaseService), stageHandler.GetStages)
			stages.POST("", stageHandler.CreateStage)
			stages.POST("/sync", auth.FirebaseAuth(app.FirebaseService), stageHandler.SyncStages)
			stages.PUT("/:stageNo/clear", auth.OptionalFirebaseAuth(app.FirebaseService), stageHandler.ClearStage)
		}

		users := v2.Group("/users")
		{
			users.POST("/login", stageHandler.Login)
			users.DELETE("/delete-account", auth.FirebaseAuth(app.FirebaseService), stageHandler.DeleteAccount)
		}
	}

	return router
}
