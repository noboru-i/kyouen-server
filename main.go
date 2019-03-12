package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
)

type KyouenPuzzleSummary struct {
	Count    int
	LastDate time.Time
}

type Statics struct {
	Count    int       `json:"count"`
	LastDate time.Time `json:"last_updated_at"`
}

func main() {
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := context.Background()
	projectID := "my-android-server"
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}

	var entities []KyouenPuzzleSummary
	q := datastore.NewQuery("KyouenPuzzleSummary").Limit(1)
	if _, err := client.GetAll(ctx, q, &entities); err != nil {
		fmt.Fprintf(w, "error! : %v", err)
		return
	}

	statics := Statics{Count: entities[0].Count, LastDate: entities[0].LastDate}
	json.NewEncoder(w).Encode(statics)
}
