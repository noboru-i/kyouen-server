package handlers

import (
	"context"
	"encoding/json"
	"kyouen-server/db"
	"kyouen-server/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyouen-server/openapi"

	"cloud.google.com/go/datastore"
	firebase "firebase.google.com/go"

	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

func ClearHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stageNo, err := strconv.Atoi(mux.Vars(r)["stageNo"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("stageNo: %v", stageNo)

	var param openapi.ClearStage
	err = json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("param: %v", param)
	size := int(math.Sqrt(float64(len(param.Stage))))
	paramKyouenStage := models.NewKyouenStage(size, param.Stage)

	if !isKyouen(paramKyouenStage) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var entities []db.KyouenPuzzle
	q := datastore.NewQuery("KyouenPuzzle").Filter("stageNo =", stageNo).Limit(1)
	keys, err := db.DB().GetAll(ctx, q, &entities)
	if err != nil {
		log.Printf("failed to get KyouenPuzzle. " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(entities) == 0 {
		log.Printf("entities is empty.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stageKey := keys[0]
	stage := entities[0]
	log.Printf("stage: %v", stage)

	if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
		log.Printf("stage is wrong.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userKey, user := getUser(r)
	log.Printf("user: %v", user)

	saveStageUser(stageKey, userKey, user)
}

func isKyouen(kyouenStage *models.KyouenStage) bool {
	return kyouenStage.IsKyouenByWhite() != nil
}

func getUser(r *http.Request) (*datastore.Key, *db.User) {
	ctx := context.Background()
	idToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	log.Printf("idToken: %v", idToken)
	if len(idToken) == 0 {
		log.Fatalf("error idToken is needed.")
	}

	opt := option.WithCredentialsFile("api-project-1046368181881-firebase-adminsdk-df1u6-039d87ad7a.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
	}

	key, user := findUserByUserID(token.UID)

	return key, user
}

func findUserByUserID(id string) (*datastore.Key, *db.User) {
	ctx := context.Background()

	var users []db.User
	q := datastore.NewQuery("User").Filter("userId =", id).Limit(1)
	keys, err := db.DB().GetAll(ctx, q, &users)
	if err != nil || len(users) == 0 {
		log.Fatalf("error getting user. id: %v, error: %v\n", id, err)
	}

	return keys[0], &users[0]
}

func saveStageUser(stageKey *datastore.Key, userKey *datastore.Key, user *db.User) {
	ctx := context.Background()

	if userKey == nil {
		// create guest user
		user = &db.User{
			UserID:          "0",
			ClearStageCount: 0,
			ScreenName:      "Guest",
			Image:           "http://kyouen.app/image/icon.png",
		}
		userKey := datastore.NameKey("User", "KEY"+"0", nil)
		if _, err := db.DB().Put(ctx, userKey, &user); err != nil {
			panic("database error." + err.Error())
		}
	}

	// check stored StageUser
	var stageUsers []db.StageUser
	q := datastore.NewQuery("StageUser").Filter("stage =", stageKey).Filter("user =", userKey).Limit(1)
	stageUserKeys, err := db.DB().GetAll(ctx, q, &stageUsers)
	if err != nil {
		log.Fatalf("failed to get StageUser. %v\n", err)
	}

	if len(stageUsers) == 0 {
		// user solve stage first time
		stageUser := db.StageUser{
			StageKey:  stageKey,
			UserKey:   userKey,
			ClearDate: time.Now(),
		}
		key := datastore.IncompleteKey("StageUser", nil)
		if _, err := db.DB().Put(ctx, key, &stageUser); err != nil {
			log.Fatalf("database error: %v", err.Error())
		}

		user.ClearStageCount++
		if _, err := db.DB().Put(ctx, userKey, &user); err != nil {
			log.Fatalf("database error: %v", err.Error())
		}
	} else {
		stageUser := stageUsers[0]
		stageUserKey := stageUserKeys[0]
		stageUser.ClearDate = time.Now()

		if _, err := db.DB().Put(ctx, stageUserKey, &stageUser); err != nil {
			log.Fatalf("database error: %v", err.Error())
		}
	}
}
