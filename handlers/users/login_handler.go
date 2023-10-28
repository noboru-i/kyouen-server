package users

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"kyouen-server/db"
	"kyouen-server/openapi"

	"cloud.google.com/go/datastore"
	firebase "firebase.google.com/go"
	"github.com/ChimeraCoder/anaconda"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var param openapi.LoginParam
	err := json.NewDecoder(r.Body).Decode(&param)

	twitterAPI := getTwitterAPI(r, param.Token, param.TokenSecret)
	v := url.Values{}
	user, err := twitterAPI.GetSelf(v)
	if err != nil {
		log.Printf("auth error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser := upsertUser(&user, &param)

	token := getFirebaseAuthenticationToken(&dbUser)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openapi.LoginResult{
		ScreenName: user.ScreenName,
		Token:      token,
	})
}

func getTwitterAPI(r *http.Request, token string, tokenSecret string) *anaconda.TwitterApi {
	return anaconda.NewTwitterApiWithCredentials(
		token,
		tokenSecret,
		os.Getenv("CONSUMER_KEY"),
		os.Getenv("CONSUMER_SECRET"))
}

func getFirebaseAuthenticationToken(user *db.User) string {
	ctx := context.Background()
	opt := option.WithCredentialsFile("api-project-1046368181881-firebase-adminsdk-df1u6-039d87ad7a.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
		panic("error initializing app. " + err.Error())
	}
	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
		panic("error getting Auth client. " + err.Error())
	}

	token, err := client.CustomToken(ctx, user.UserID)
	if err != nil {
		log.Fatalf("error minting custom token: %v", err)
		panic("error minting custom token. " + err.Error())
	}

	log.Printf("Got custom token: %v", token)
	return token
}

func upsertUser(user *anaconda.User, param *openapi.LoginParam) db.User {
	ctx := context.Background()
	key := datastore.NameKey("User", "KEY"+strconv.FormatInt(user.Id, 10), nil)
	var dbUser db.User
	err := db.DB().Get(ctx, key, &dbUser)
	if err != nil && err != datastore.ErrNoSuchEntity {
		panic("database error. " + err.Error())
	}
	if err == datastore.ErrNoSuchEntity {
		log.Printf("new user. name=%s", user.ScreenName)
		dbUser = db.User{
			UserID:          fmt.Sprint(user.Id),
			ClearStageCount: 0,
		}
	}
	dbUser.ScreenName = user.ScreenName
	dbUser.Image = user.ProfileImageUrlHttps
	dbUser.AccessToken = param.Token
	dbUser.AccessSecret = param.TokenSecret

	if _, err := db.DB().Put(ctx, key, &dbUser); err != nil {
		panic("database error." + err.Error())
	}

	return dbUser
}
