package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"kyouen-server/db"
	"kyouen-server/openapi"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
)

func StaticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	projectID := "my-android-server"
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}

	var entities []db.KyouenPuzzleSummary
	q := datastore.NewQuery("KyouenPuzzleSummary").Limit(1)
	if _, err := client.GetAll(ctx, q, &entities); err != nil {
		fmt.Fprintf(w, "error! : %v", err)
		return
	}

	statics := openapi.Statics{Count: entities[0].Count, LastUpdatedAt: entities[0].LastDate}
	json.NewEncoder(w).Encode(statics)
}
