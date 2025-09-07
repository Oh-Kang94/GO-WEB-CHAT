package main

import (
	"log"
	"net/http"
	"os"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

const (
	nextPageKey = "next_page"
)

func init() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env 파일을 찾을 수 없음, 시스템 환경변수 사용")
	}

	// 환경변수 읽기
	authSecurityKey := os.Getenv("AUTH_SECURITY_KEY")
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	// gomniauth 초기화
	gomniauth.SetSecurityKey(authSecurityKey)
	gomniauth.WithProviders(
		google.New(clientID, clientSecret, redirectURL),
	)
}

func loginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	action := ps.ByName("action")
	provider := ps.ByName("provider")

	s := sessions.GetSession(r)

	switch action {
	case "login":
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}

		loginUrl, err := p.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln(err)
		}

		http.Redirect(w, r, loginUrl, http.StatusFound)
	case "callback":
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}

		creds, err := p.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln(err)
		}

		user, err := p.GetUser(creds)

		if err != nil {
			log.Fatalln(err)
		}

		u := &User{
			Uid:       user.Data().Get("id").MustStr(),
			Name:      user.Name(),
			Email:     user.Email(),
			AvatarUrl: user.AvatarURL(),
		}

		SetCurrentUser(r, u)

		http.Redirect(w, r, s.Get(nextPageKey).(string), http.StatusFound)
	default:
		http.Error(w, "Auth action is Not Support", http.StatusNotFound)
	}

}
