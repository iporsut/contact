package contact

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var session *mgo.Session

type MongoProvider struct {
	session *mgo.Session
}

func NewMongoProvider() *MongoProvider {
	session, _ := mgo.Dial("localhost")
	return &MongoProvider{session}
}

func CloneSession() *mgo.Session {
	s := session.Clone()
	s.SetMode(mgo.Monotonic, true)
	return s
}

func ContactCollection(s *mgo.Session) *mgo.Collection {
	return s.DB("test").C("contact")
}

type queryFunc func(c *mgo.Collection) error

func doQuery(session *mgo.Session, query queryFunc) error {
	s := session.Clone()
	s.SetMode(mgo.Monotonic, true)
	defer s.Close()
	c := ContactCollection(s)
	return query(c)
}

func (mp *MongoProvider) Get(id string) (result Information, err error) {
	get := func(c *mgo.Collection) error {
		return c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	}

	err = doQuery(mp.session, get)
	err = handleError(err)

	return
}

func (mp *MongoProvider) All() []Information {
	var result []Information
	all := func(c *mgo.Collection) error {
		return c.Find(nil).All(&result)
	}

	doQuery(mp.session, all)

	return result
}

func (mp *MongoProvider) Update(i Information) error {
	err := sessionHandler(update)(i)
	return handleError(err)
}

func (mp *MongoProvider) Delete(id string) error {
	err := sessionHandler(delete)(id)
	return handleError(err)
}

func (mp *MongoProvider) Add(i *Information) error {
	err := sessionHandler(add)(i)
	return handleError(err)
}

type action func(*mgo.Collection, interface{}) error
type wrapper func(interface{}) error

func sessionHandler(a action) wrapper {
	return func(input interface{}) error {
		s := CloneSession()
		c := ContactCollection(s)
		defer s.Close()

		return a(c, input)
	}
}

func update(c *mgo.Collection, input interface{}) error {
	i, _ := input.(Information)
	target := bson.M{"id": i.Id}
	change := bson.M{"$set": bson.M{"id": i.Id, "email": i.Email, "title": i.Title, "content": i.Content}}
	return c.Update(target, change)
}

func delete(c *mgo.Collection, input interface{}) error {
	id, _ := input.(string)
	target := bson.M{"id": id}
	return c.Remove(target)
}

func add(c *mgo.Collection, input interface{}) error {
	i, _ := input.(Information)
	return c.Insert(i)
}

func handleError(err error) error {
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
