package session

import (
	"fmt"
	"github.com/go-martini/martini"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testSession struct {
	Uid         bson.ObjectId `bson:"_id,omitempty"`
	CreateTime int64
}

func (s *testSession) Id() string {
	return s.Uid.Hex()
}

func (s *testSession) IsOutOfDate() bool {
	return false
}

func newTestSession() Session {
	return &testSession{bson.NewObjectId(), time.Now().Unix()}
}

type testHttpServer struct {
	*martini.ClassicMartini
	Store
	sessionId string
}

func (s *testHttpServer) Get(path string, cb func(res *httptest.ResponseRecorder)) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	if s.Store.Len() > 0 {
		req.AddCookie(&http.Cookie{Name: CookieName, Value: s.sessionId})
	}
	s.ServeHTTP(res, req)
	cb(res)
}

func newTestHttpServer(store Store) *testHttpServer {
	sessionId := bson.NewObjectId()
	m := martini.Classic()
	m.Use(Midware(store))

	m.Get("/login", func(s Store) {
		s.Add(&testSession{sessionId, time.Now().Unix()})
	})

	m.Get("/logout", Auth(), func(s Store, session Session) {
		s.Remove(session.Id())
	})

	m.Get("/private", Auth(), func() string { return "ok" })
	return &testHttpServer{m, store, sessionId.Hex()}
}

func TestMemoryStore_Add(t *testing.T) {
	ts := newTestSession()
	store := NewMemoryStore()
	b := store.Add(ts)
	if !b {
		t.Error("Add fail!")
	}

	if store.Len() != 1 {
		t.Error("Store len should be one!")
	}

	b = store.Add(ts)
	if b {
		t.Error("Should add fail")
	}

	if store.Len() != 1 {
		t.Error("Should be one!")
	}
}

func auth(t *testing.T, s *testHttpServer) {
	s.Get("/private", func(res *httptest.ResponseRecorder) {
		if res.Code == http.StatusUnauthorized {
			return
		}
		t.Errorf("Status code should be %d but was %d!", http.StatusUnauthorized, res.Code)
	})

	s.Get("/login", func(res *httptest.ResponseRecorder) {
		setcookie := res.Header().Get("Set-Cookie")
		if setcookie != "" {
			fmt.Printf("Set cookie header %s!\n", setcookie)
			return
		}
		t.Error("Response have no set cookie header!")

		if s.Store.Has(s.sessionId) {
			return
		}
		t.Error("SessionId not be saved in session store!")
	})

	s.Get("/private", func(res *httptest.ResponseRecorder) {
		if res.Code == http.StatusOK {
			return
		}
		t.Errorf("Status code should be %d but was %d!", http.StatusOK, res.Code)
	})

	s.Get("/logout", func(res *httptest.ResponseRecorder) {
		if !s.Store.Has(s.sessionId) {
			return
		}
		t.Error("Store should be remove sessionid!")
	})

	s.Get("/private", func(res *httptest.ResponseRecorder) {
		if res.Code == http.StatusUnauthorized {
			return
		}
		t.Errorf("Status code should be %d but was %d!", http.StatusUnauthorized, res.Code)
	})
}

func TestAuthForMemoryStore(t *testing.T) {
	fmt.Println("Test auth for memory store!")
	auth(t, newTestHttpServer(NewMemoryStore()))
}

func TestAuthForMongoStore(t *testing.T) {
	fmt.Println("Test Auth for mongo store!")
	mgoSession, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Error(err)
	}
	defer mgoSession.Close()
	mgoSession.SetMode(mgo.Monotonic, true)
	c := mgoSession.DB("TestSession").C("sessions")
	store := NewMongoStore(c, func() Session {
		return &testSession{}
	})
	auth(t, newTestHttpServer(store))
}
