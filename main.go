package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"kyouen-server/db"
	"kyouen-server/handlers"
	usersHandler "kyouen-server/handlers/users"

	"google.golang.org/appengine"
)

func main() {
	db.InitDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/users/login", usersHandler.LoginHandler)

	mux.HandleFunc("/v2/stages", handlers.StagesHandler)

	mux.HandleFunc("/v2/statics", handlers.StaticsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	filteredHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		mux.ServeHTTP(w, r)
	})
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), filteredHandler))
}
