package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"kyouen-server/db"
	"kyouen-server/handlers"
	stageNoHandler "kyouen-server/handlers/stages/stage_no"
	usersHandler "kyouen-server/handlers/users"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
)

func main() {
	db.InitDB()

	r := mux.NewRouter().PathPrefix("/v2").Subrouter()

	r.HandleFunc("/users/login", usersHandler.LoginHandler)

	r.HandleFunc("/stages", handlers.StagesHandler)
	r.HandleFunc("/stages/{stageNo}/clear", stageNoHandler.ClearHandler)

	r.HandleFunc("/statics", handlers.StaticsHandler)

	r.Use(corsMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if appengine.IsDevAppServer() {
			// allow executing API from Swagger UI
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
