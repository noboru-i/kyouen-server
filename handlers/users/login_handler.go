package users

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"kyouen-server/openapi"

	"github.com/ChimeraCoder/anaconda"
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openapi.LoginResult{ScreenName: user.ScreenName})
}

func getTwitterAPI(r *http.Request, token string, tokenSecret string) *anaconda.TwitterApi {
	return anaconda.NewTwitterApiWithCredentials(
		token,
		tokenSecret,
		os.Getenv("CONSUMER_KEY"),
		os.Getenv("CONSUMER_SECRET"))
}
