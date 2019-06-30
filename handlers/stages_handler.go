package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"kyouen-server/db"
	"kyouen-server/models"
	"kyouen-server/openapi"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

func StagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		stagesGetHandler(w, r)
	case http.MethodPost:
		stagesPostHandler(w, r)
	}
}

func stagesGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	param := parseGetParam(r)

	var entities []db.KyouenPuzzle
	q := datastore.NewQuery("KyouenPuzzle").Filter("stageNo >=", param.startStageNo).Limit(param.limit)
	if _, err := db.DB().GetAll(ctx, q, &entities); err != nil {
		fmt.Fprintf(w, "error! : %v", err)
		return
	}

	var stageList []openapi.Stage
	for _, value := range entities {
		stageList = append(stageList, openapi.Stage{
			StageNo:    value.StageNo,
			Size:       value.Size,
			Stage:      value.Stage,
			Creator:    value.Creator,
			RegistDate: value.RegistDate})
	}
	json.NewEncoder(w).Encode(stageList)
}

type getParam struct {
	startStageNo int
	limit        int
}

func parseGetParam(r *http.Request) getParam {
	startStageNo, err := strconv.Atoi(r.FormValue("start_stage_no"))
	if err != nil {
		startStageNo = 0
	}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return getParam{startStageNo: startStageNo, limit: limit}
}

func stagesPostHandler(w http.ResponseWriter, r *http.Request) {
	var param openapi.NewStage
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stage := *models.NewKyouenStage(int(param.Size), param.Stage)

	// check stone count
	if stage.StoneCount() <= 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "stage must have 5 stones.")
		return
	}

	// check stones is kyouen
	kyouenData := stage.HasKyouen()
	if kyouenData == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "sent stage don't have kyouen.")
		return
	}

	// check registered
	if hasRegisteredStageAll(stage) {
		// TODO change result
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "sent stage is already exists.")
		return
	}

	// save to datastore
	savedStage := saveStage(param, getNextStageNo())

	// create response
	json.NewEncoder(w).Encode(savedStage)
}

func getNextStageNo() int64 {
	ctx := context.Background()

	var entities []db.KyouenPuzzle
	q := datastore.NewQuery("KyouenPuzzle").Order("-stageNo").Limit(1)
	if _, err := db.DB().GetAll(ctx, q, &entities); err != nil {
		return 1
	}
	if len(entities) == 0 {
		return 1
	}

	return entities[0].StageNo + 1
}

func hasRegisteredStage(stage string) bool {
	ctx := context.Background()

	q := datastore.NewQuery("KyouenPuzzle").Filter("stage =", stage).Limit(1)
	count, err := db.DB().Count(ctx, q)
	if err != nil {
		panic("database error." + err.Error())
	}
	return count != 0
}

func hasRegisteredStageAll(stage models.KyouenStage) bool {
	for i := 0; i < 4; i++ {
		mirror := models.NewMirroredKyouenStage(stage)
		if hasRegisteredStage(mirror.ToString()) {
			return true
		}

		stage = *models.NewRotatedKyouenStage(stage)
		if hasRegisteredStage(stage.ToString()) {
			return true
		}
	}

	return false
}

func saveStage(param openapi.NewStage, newStageNo int64) openapi.Stage {
	ctx := context.Background()

	stage := db.KyouenPuzzle{
		StageNo:    newStageNo,
		Size:       param.Size,
		Stage:      param.Stage,
		Creator:    param.Creator,
		RegistDate: time.Now(),
	}
	key := datastore.IncompleteKey("KyouenPuzzle", nil)
	if _, err := db.DB().Put(ctx, key, &stage); err != nil {
		panic("database error." + err.Error())
	}

	increaseSummaryCount()

	return openapi.Stage{
		StageNo:    stage.StageNo,
		Size:       stage.Size,
		Stage:      stage.Stage,
		Creator:    stage.Creator,
		RegistDate: stage.RegistDate,
	}
}

func increaseSummaryCount() {
	ctx := context.Background()

	k := datastore.IDKey("KyouenPuzzleSummary", 1, nil)
	var summary db.KyouenPuzzleSummary
	err := db.DB().Get(ctx, k, &summary)
	if err != nil {
		q := datastore.NewQuery("KyouenPuzzle").KeysOnly()
		count, err := db.DB().Count(ctx, q)
		if err != nil {
			panic("database error." + err.Error())
		}
		summary = db.KyouenPuzzleSummary{
			Count:    int64(count),
			LastDate: time.Now(),
		}
	} else {
		summary.Count++
	}

	_, err = db.DB().Put(ctx, k, &summary)
	if err != nil {
		panic("database error." + err.Error())
	}
}
