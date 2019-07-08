package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func ClearHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "stageNo: %v\n", vars["stageNo"])

	// ctx := context.Background()

	// var entities []db.KyouenPuzzleSummary
	// q := datastore.NewQuery("KyouenPuzzleSummary").Limit(1)
	// if _, err := db.DB().GetAll(ctx, q, &entities); err != nil {
	// 	fmt.Fprintf(w, "error! : %v", err)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	// statics := openapi.Statics{Count: entities[0].Count, LastUpdatedAt: entities[0].LastDate}
	// json.NewEncoder(w).Encode(statics)
}
