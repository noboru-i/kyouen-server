package handlers

import (
	"context"
	"encoding/json"
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

type GetParam struct {
	startStageNo int
	limit        int
}

func parseGetParam(r *http.Request) GetParam {
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

	return GetParam{startStageNo: startStageNo, limit: limit}
}

func stagesPostHandler(w http.ResponseWriter, r *http.Request) {
	// newStageNo := getNextStageNo()
	// w.Write([]byte(strconv.FormatInt(newStageNo, 10)))

	// result := hasRegisteredStage("000000001010000000010010000000001010")
	// if result {
	// 	w.Write([]byte("ok"))
	// } else {
	// 	w.Write([]byte("ng"))
	// }

	// result := rotateMatrix("000000001010000000010010000000001010", 6)
	// w.Write([]byte(result))
	// w.Write([]byte("\n"))
	// result2 := mirrorMatrix("000000001010000000010010000000001010", 6)
	// w.Write([]byte(result2))

	result := hasRegisteredStageAll("000000001010000000010010000000001010")
	w.Write([]byte(strconv.FormatBool(result)))

	// # POSTリクエストを処理します。
	// def post(self):
	//     # パラメータ名：dataを取得
	//     data = self.request.get('data').split(',')
	//     logging.debug("post data:" + str(data))

	//     if len(data) != 3:
	//         # 要素が3つ取得できない場合はエラー
	//         self.response.headers['Content-Type'] = 'text/plain'
	//         self.response.out.write('error' + str(data))
	//         return

	//     from kyouenmodule import hasKyouen, getPoints
	//     if len(getPoints(data[1], int(data[0]))) <= 4:
	//         # 石の数が4以下の場合
	//         self.response.headers['Content-Type'] = 'text/plain'
	//         self.response.out.write("not enough stone")
	//         logging.error('not enough stone.' + data[1])
	//         return
	//     if not hasKyouen(getPoints(data[1], int(data[0]))):
	//         # 共円でない場合
	//         self.response.headers['Content-Type'] = 'text/plain'
	//         self.response.out.write("not kyouen")
	//         logging.error('not kyouen.' + data[1])
	//         return

	//     if self.checkRegistered(data[1], int(data[0])):
	//         # 登録済みの場合
	//         self.response.headers['Content-Type'] = 'text/plain'
	//         self.response.out.write("registered")
	//         return

	//     # 入力データの登録
	//     model = KyouenPuzzle(stageNo=self.getNextStageNo(),
	//                          size=int(data[0]),
	//                          stage=data[1],
	//                          creator=data[2].replace('\n', ''))
	//     model.put()

	//     # DB登録
	//     regist_model = RegistModel(stageInfo=model.key)
	//     regist_model.put()

	//     # サマリの再計算
	//     summary = KyouenPuzzleSummary.query().get()
	//     if not summary:
	//         c = KyouenPuzzle.query().count()
	//         summary = KyouenPuzzleSummary(count=c)
	//     else:
	//         summary.count += 1
	//     summary.put()

	//     # レスポンスの返却
	//     self.response.headers['Content-Type'] = 'text/plain'
	//     self.response.out.write('success stageNo=' + str(model.stageNo))
	//     return
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
