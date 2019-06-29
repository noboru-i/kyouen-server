package users

import (
	"fmt"
	"net/http"
	"os"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	v := os.Getenv("CONSUMER_KEY")
	fmt.Fprintf(w, "env var is \"%v\".", v)
}

type postParam struct {
	token       string
	tokenSecret string
}

func parsePostParam(r *http.Request) postParam {
	token := r.FormValue("token")
	tokenSecret := r.FormValue("token_secret")

	return postParam{token: token, tokenSecret: tokenSecret}
}
