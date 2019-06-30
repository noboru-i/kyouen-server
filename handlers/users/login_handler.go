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

	param := parsePostParam(r)
	twitterAPI := getTwitterAPI(r, param.Token, param.TokenSecret)
	v := url.Values{}
	user, err := twitterAPI.GetSelf(v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(openapi.LoginResult{ScreenName: user.ScreenName})
}

func parsePostParam(r *http.Request) openapi.LoginParam {
	token := r.FormValue("token")
	tokenSecret := r.FormValue("token_secret")

	return openapi.LoginParam{Token: token, TokenSecret: tokenSecret}
}

func getTwitterAPI(r *http.Request, token string, tokenSecret string) *anaconda.TwitterApi {
	return anaconda.NewTwitterApiWithCredentials(
		token,
		tokenSecret,
		os.Getenv("CONSUMER_KEY"),
		os.Getenv("CONSUMER_SECRET"))
}
