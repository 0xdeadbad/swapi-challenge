package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"swapi-challenge/client"
	"time"

	"strings"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Person struct {
	ID     primitive.ObjectID `json:"_id,omitempty" redis:"_id"    bson:"_id"`
	Pid    int                `json:"id,omitempty"  redis:"id"     bson:"id"`
	Height string             `json:"height"        redis:"height" bson:"height"`
	Name   string             `json:"name"          redis:"name"   bson:"name"`
	Gender string             `json:"gender"        redis:"gender" bson:"gender"`
}

type SearchResult struct {
	Count   int      `json:"count"`
	Results []Person `json:"results"`
}

type PeopleAPIEndpoints struct {
	coll        *mongo.Collection
	redisClient *redis.Client
}

func NewPeopleAPIEndpoints(r *mux.Router, redisClient *redis.Client, mongoDatabase *mongo.Database) (*PeopleAPIEndpoints, error) {
	coll := mongoDatabase.Collection("people")

	p := &PeopleAPIEndpoints{
		coll:        coll,
		redisClient: redisClient,
	}

	r.Path("/{name}").HandlerFunc(p.getPersonEndpoint).Methods("GET")

	return p, nil
}

func (p *PeopleAPIEndpoints) getPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	name := mux.Vars(req)["name"]

	person := Person{}
	err := p.cachedQuery(req.Context(), name, &person)
	if err == ErrPersonNotFound {
		err := p.netQuery(req.Context(), name, &person)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		person.ID = primitive.NewObjectID()
		err = p.cachedInsert(req.Context(), person)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}
	}

	d, err := json.Marshal(person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.Write(d)
}

func (p *PeopleAPIEndpoints) cachedInsert(ctx context.Context, person Person) error {

	_, err := p.coll.InsertOne(ctx, &person)
	if err != nil {
		return err
	}

	//_, err = p.redisClient.JSONSet(ctx, person.Name, "$", &person).Result()
	_, err = p.redisClient.HSet(ctx, fmt.Sprintf("STRUCT:%s", strings.ReplaceAll(person.Name, " ", "")), person).Result()
	if err != nil {
		return err
	}

	return nil
}

func (p *PeopleAPIEndpoints) cachedQuery(ctx context.Context, name string, person *Person) error {

	// cmd := p.redisClient.JSONGet(ctx, name, "$")
	per, err := p.redisClient.HGetAll(ctx, fmt.Sprintf("STRUCT:%s", strings.ReplaceAll(name, " ", ""))).Result()
	if err != nil {
		return err
	}

	j, err := json.Marshal(&per)
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	if string(j) == "" {
		per := Person{}
		err := p.coll.FindOne(ctx, &Person{Name: name}).Decode(&per)
		if err != nil {
			return ErrPersonNotFound
		}
		*person = per

		_, err = p.redisClient.HSet(ctx, fmt.Sprintf("STRUCT:%s", strings.ReplaceAll(name, " ", "")), person).Result()
		if err != nil {
			return err
		}

		return nil
	}

	err = json.Unmarshal(j, person)
	if err != nil {
		return err
	}

	return nil
}

func (p *PeopleAPIEndpoints) netQuery(ctx context.Context, value string, person *Person) error {
	httpClient, err := client.NewHTTPClient(client.WithTimeout(time.Duration(time.Second * 20)))
	if err != nil {
		return err
	}

	resp, err := httpClient.Get(fmt.Sprintf("https://swapi.dev/api/people/?search=%s&format=json", value))
	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	results := SearchResult{}
	err = json.Unmarshal(data, &results)
	if err != nil {
		return err
	}

	*person = results.Results[0]

	return nil
}

type PersonReqError string

const (
	ErrPersonNotFound PersonReqError = "person requested not found"
)

func (e PersonReqError) Error() string {
	return string(e)
}

func (e PersonReqError) String() string {
	return string(e)
}
