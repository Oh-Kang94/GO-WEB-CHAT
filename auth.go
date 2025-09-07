package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/urfave/negroni"
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
			log.Printf("Error getting provider: %v", err)
			http.Error(w, "Authentication provider error", http.StatusInternalServerError)
			return
		}

		loginUrl, err := p.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Printf("Error getting auth URL: %v", err)
			http.Error(w, "Authentication URL error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, loginUrl, http.StatusFound)

	case "callback":
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Printf("Error getting provider: %v", err)
			http.Error(w, "Authentication provider error", http.StatusInternalServerError)
			return
		}

		creds, err := p.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Printf("Error completing auth: %v", err)
			http.Redirect(w, r, "/login?error=auth_failed", http.StatusFound)
			return
		}

		user, err := p.GetUser(creds)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			http.Redirect(w, r, "/login?error=user_fetch_failed", http.StatusFound)
			return
		}

		u := &User{
			Uid:       user.Data().Get("id").MustStr(),
			Name:      user.Name(),
			Email:     user.Email(),
			AvatarUrl: user.AvatarURL(),
		}

		// Debug log to see if user is created properly
		log.Printf("User authenticated: %+v", u)

		SetCurrentUser(r, u)

		// Handle redirect URL safely
		var redirectURL string
		if nextPage := s.Get(nextPageKey); nextPage != nil {
			if nextPageStr, ok := nextPage.(string); ok {
				redirectURL = nextPageStr
			} else {
				redirectURL = "/" // Default fallback
			}
		} else {
			redirectURL = "/" // Default fallback
		}

		// Clear the next page from session
		s.Delete(nextPageKey)

		log.Printf("Redirecting authenticated user to: %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)

	default:
		http.Error(w, "Auth action is not supported", http.StatusNotFound)
	}
}

func LoginRequired(ignore ...string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		for _, s := range ignore {
			if strings.HasPrefix(r.URL.Path, s) {
				next(w, r)
				return
			}
		}

		u := GetCurrentUser(r)

		if u != nil && u.Valid() {
			SetCurrentUser(r, u)
			next(w, r)
			return
		}

		SetCurrentUser(r, nil)

		sessions.GetSession(r).Set(nextPageKey, r.URL.RequestURI())

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
