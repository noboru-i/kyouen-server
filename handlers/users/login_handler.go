package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	param := parsePostParam(r)
	twitterAPI := getTwitterAPI(r, param.token, param.tokenSecret)
	v := url.Values{}
	user, err := twitterAPI.GetSelf(v)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %v", err)
		return
	}
	json.NewEncoder(w).Encode(postResponse{ScreenName: user.ScreenName})
}

type postParam struct {
	token       string
	tokenSecret string
}

func parsePostParam(r *http.Request) postParam {
	token := r.FormValue("token")
	tokenSecret := r.FormValue("tokenSecret")

	return postParam{token: token, tokenSecret: tokenSecret}
}

type postResponse struct {
	ScreenName string `json:"screenName"`
}

func getTwitterAPI(r *http.Request, token string, tokenSecret string) *anaconda.TwitterApi {
	return anaconda.NewTwitterApiWithCredentials(
		token,
		tokenSecret,
		os.Getenv("CONSUMER_KEY"),
		os.Getenv("CONSUMER_SECRET"))
}
