package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"kyouen-server/db"
	"kyouen-server/handlers"
)

func main() {
	db.InitDB()
	http.HandleFunc("/statics", handlers.StaticsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
