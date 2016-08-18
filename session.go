package session

import (
	"github.com/go-martini/martini"
	"net/http"
)

type Session interface {
	Id() string
	IsOutOfDate() bool
}

type Store interface {
	Has(key string) bool
	Get(key string) Session
	Add(value Session) bool
	Len() int
	Remove(key string)
}

type storeWapper struct {
	Store
	http.ResponseWriter
}

func (ws *storeWapper) Add(value Session) bool {
	http.SetCookie(ws, &http.Cookie{Name: CookieName, Value: value.Id()})
	return ws.Store.Add(value)
}

var CookieName = "SessionId"

func Midware(s Store) martini.Handler  {
	return func(res http.ResponseWriter, c martini.Context) {
		c.MapTo(&storeWapper{s, res}, (*Store)(nil))
	}
}

func Auth() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context, s Store) {
		cookie, err := req.Cookie(CookieName)
		if err != nil || !s.Has(cookie.Value) {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.Map(s.Get(cookie.Value))
	}
}
