package main

import (
	"context"
	"log"
	"net/http"
	"time"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	renderer *render.Render
	client   *mongo.Client
)

const (
	sessionKey    = "GO-WEB-CHAT"
	sessionSecret = "GO-WEB-CHAT_SECRET"
)

func init() {
	renderer = render.New()
	log.Println("MongoDB 연결 시작")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB 클라이언트 생성
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:rootpw@localhost:27017/admin"))
	if err != nil {
		log.Fatalf("MongoDB 연결 실패: %v", err)
	}

	// 연결 확인
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping 실패: %v", err)
	}

	log.Println("MongoDB 연결 성공")
}

func main() {
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		renderer.HTML(w, http.StatusOK, "index", map[string]string{"title": "Simple Chat"})
	})

	// LOGIN
	router.GET("/login", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		renderer.HTML(w, http.StatusOK, "login", nil)
	})

	router.GET("/logout", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sessions.GetSession(r).Delete(currentUserKey)
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	router.GET("/auth/:action/:provider", loginHandler)

	n := negroni.Classic()
	store := cookiestore.New([]byte(sessionSecret))
	n.Use(sessions.Sessions(sessionKey, store))
	n.Use(LoginRequired("/login", "/auth"))
	n.UseHandler(router)

	n.Run(":3000")
}
