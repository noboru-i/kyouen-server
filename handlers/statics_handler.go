package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"kyouen-server/db"
	"kyouen-server/openapi"
	"net/http"

	"cloud.google.com/go/datastore"
)

func StaticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var entities []db.KyouenPuzzleSummary
	q := datastore.NewQuery("KyouenPuzzleSummary").Limit(1)
	if _, err := db.DB().GetAll(ctx, q, &entities); err != nil {
		fmt.Fprintf(w, "error! : %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	statics := openapi.Statics{Count: entities[0].Count, LastUpdatedAt: entities[0].LastDate}
	json.NewEncoder(w).Encode(statics)
}
