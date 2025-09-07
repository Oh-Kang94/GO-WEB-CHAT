package main

import (
	"net/http"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var renderer *render.Render

const (
	sessionKey    = "GO-WEB-CHAT"
	sessionSecret = "GO-WEB-CHAT_SECRET"
)

func init() {
	renderer = render.New()
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
