package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kyouen-server/db"
	"kyouen-server/openapi"
	"math"
	"net/http"
	"strconv"
	"strings"

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
	param, err := parsePostParam(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO check stone count

	// TODO check stones is kyouen

	// TODO check registered
	result := hasRegisteredStageAll(param.stage)
	if result {
		w.Write([]byte("registered"))
	} else {
		w.Write([]byte("not registered"))
	}

	// TODO register KyouenPuzzle
	// newStageNo := getNextStageNo()
	// w.Write([]byte(strconv.FormatInt(newStageNo, 10)))

	// TODO update KyouenPuzzleSummary

	// TODO create response
}

type postParam struct {
	size    int
	stage   string
	creator string
}

func parsePostParam(r *http.Request) (postParam, error) {
	size, err := strconv.Atoi(r.FormValue("size"))
	if err != nil {
		return postParam{}, errors.New("size is invalid")
	}
	stage := r.FormValue("stage")
	creator := r.FormValue("creator")

	return postParam{size: size, stage: stage, creator: creator}, nil
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
		panic("database error.")
	}
	return count != 0
}

func hasRegisteredStageAll(stage string) bool {
	if hasRegisteredStage(stage) {
		return true
	}

	size := int(math.Sqrt(float64(len(stage))))

	rotate := stage
	for i := 0; i < 3; i++ {
		rotate = rotateMatrix(rotate, size)
		if hasRegisteredStage(rotate) {
			return true
		}
	}

	mirror := mirrorMatrix(stage, size)
	if hasRegisteredStage(mirror) {
		return true
	}

	return false
}

// TODO extract to other file.

func rotateMatrix(stage string, size int) string {
	result := make([]string, size*size)
	for i, s := range stage {
		x := i % size
		y := i / size
		result[(size-y-1)+x*size] = string(s)
	}
	return strings.Join(result, "")
}

func mirrorMatrix(stage string, size int) string {
	result := make([]string, size*size)
	for i, s := range stage {
		x := i % size
		y := i / size
		result[(size-x-1)+y*size] = string(s)
	}
	return strings.Join(result, "")
}
