package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	. "github.com/fulldump/box"
	"github.com/globalsign/mgo"
)

func main() {

	fmt.Println("Connecting to mongodb...")
	session, err := NewMongoSession("mongodb://tfb-database/hello_world")
	if err != nil {
		panic(err)
	}

	b := NewBox()

	b.WithInterceptors(
		SetHeader("Server", "box/v0.1.3"),
		NewMongoInterceptor(session),
	)

	b.Resource("/json").WithActions(
		Get(serializeJSON),
	)

	b.Resource("/db").WithActions(
		Get(singleQuery),
	)

	b.Resource("/queries").WithActions(
		Get(multipleQueries),
	)

	b.Resource("/fortunes").WithActions(
		Get(fortunes),
	)

	b.Resource("/plaintext").WithActions(
		Get(plaintext),
	)

	b.Resource("/updates").WithActions(
		Get(dbupdate),
	)

	b.Serve()
}

func SetHeader(key, value string) I {
	return func(next H) H {
		return func(ctx context.Context) {
			w := GetResponse(ctx)
			w.Header().Set(key, value)
			next(ctx)
		}
	}
}

// GET /json
var helloWorldMessage = &struct {
	Message string `json:"message"`
}{
	Message: "Hello, World!",
}

func serializeJSON(w http.ResponseWriter) interface{} {
	w.Header().Set("Content-Type", "application/json")
	return helloWorldMessage
}

// GET /db
type World struct {
	ID           int `json:"id"`
	RandomNumber int `json:"randomnumber"`
}

func singleQuery(ctx context.Context) (w World, err error) {
	GetResponse(ctx).Header().Set("Content-Type", "application/json")
	worldID := 1 + rand.Intn(10000)
	err = GetMongoCollection(ctx, "world").FindId(worldID).One(&w)
	return
}

// GET /queries
func multipleQueries() string {
	return "multipleQueries"
}

// GET /fortunes
func fortunes() string {
	return "fortunes"
}

// GET /plaintext
var text = []byte("Hello, World!")

func plaintext(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(text)
}

// GET /updates
func dbupdate() string {
	return "dbupdate"
}

// mongo
func NewMongoSession(mongouri string) (*mgo.Session, error) {

	info, _ := mgo.ParseURL(mongouri) // TODO: handle error
	info.Timeout = 10 * time.Second
	info.FailFast = true

	return mgo.DialWithInfo(info)
}

var ContextKeySession = "30caaa14-5873-11ec-92fd-a79a59020e01"

// NewMongoInterceptor returns a box interceptor from a valid mongo session. It
// also ensure a fresh session for each request.
func NewMongoInterceptor(session *mgo.Session) I {

	return func(next H) H {

		return func(ctx context.Context) {

			// Ensure a fresh session for each request
			s := session.Clone()
			defer s.Close()

			ctx = SetMongoSession(ctx, s)
			next(ctx)
		}
	}
}

func SetMongoSession(ctx context.Context, s *mgo.Session) context.Context {
	return context.WithValue(ctx, ContextKeySession, s)
}

func GetMongoSession(ctx context.Context) *mgo.Session {

	v := ctx.Value(ContextKeySession)
	if nil == v {
		panic("Mongo session should be in context!!!!")
	}

	return v.(*mgo.Session)
}
func GetMongoCollection(ctx context.Context, collection string) *mgo.Collection {
	return GetMongoSession(ctx).DB("").C(collection)
}
