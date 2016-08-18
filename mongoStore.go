package session

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoStore struct {
	*mgo.Collection
	sessionCreater func() Session
}

func (s *mongoStore) Has(key string) bool {
	id := bson.ObjectIdHex(key)
	cnt, err := s.FindId(id).Count()
	return err == nil && cnt != 0
}

func (s *mongoStore) Get(key string) Session {
	ret := s.sessionCreater()
	id := bson.ObjectIdHex(key)
	err := s.FindId(id).One(ret)
	if err != nil {
		return nil
	}
	return ret
}

func (s *mongoStore) Add(value Session) bool {
	err := s.Insert(value)
	return err == nil
}

func (s *mongoStore) Len() int {
	cnt, _ := s.Count()
	return cnt
}

func (s *mongoStore) Remove(key string) {
	id := bson.ObjectIdHex(key)
	s.RemoveId(id)
}

func NewMongoStore(collection *mgo.Collection, sc func() Session) Store {
	return &mongoStore{collection, sc}
}
