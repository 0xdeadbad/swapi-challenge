package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	httpclient "swapi-challenge/api/client"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Person struct {
	Pid    int    `json:"id,omitempty" redis:"id"     bson:"id"`
	Height string `json:"height"       redis:"height" bson:"height"`
	Name   string `json:"name"         redis:"name"   bson:"name"`
	Gender string `json:"gender"       redis:"gender" bson:"gender"`
}

type PeopleApiHandler struct {
	coll        *mongo.Collection
	redisClient *redis.Client
}

func newPeopleApiHandler(r *mux.Router, coll *mongo.Collection, redisClient *redis.Client) {
	p := &PeopleApiHandler{
		coll:        coll,
		redisClient: redisClient,
	}

	r.Path("/").Methods("GET").HandlerFunc(p.peopleApiEndpoint)
}

func (h *PeopleApiHandler) peopleApiEndpoint(w http.ResponseWriter, req *http.Request) {

	httpClient, err := httpclient.NewHTTPClient()
	if err != nil {
		logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	qid := req.URL.Query().Get("id")

	if qid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(qid, 10, 64)
	if err != nil {
		logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redisReq := fmt.Sprintf("STRUCT:%d", id)

	person := Person{Pid: int(id)}
	err = h.redisClient.HGetAll(req.Context(), redisReq).Scan(&person)
	if err == redis.Nil || person.Gender == "" || person.Height == "" || person.Name == "" {
		logger.Log(req.Context(), slog.LevelInfo, "Request didn't hit Redis")
		res := h.coll.FindOne(req.Context(), &person)
		err = res.Err()
		if err != nil && err != mongo.ErrNoDocuments {
			logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = res.Decode(&person)
		if err == mongo.ErrNoDocuments {
			logger.Log(req.Context(), slog.LevelInfo, "Request couldn't be found on MongoDB")
			resp, err := httpClient.Get(fmt.Sprintf("https://swapi.dev/api/people/%d/?format=json", id))
			if err != nil {
				logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if resp.StatusCode != http.StatusOK {
				logger.Log(req.Context(), slog.LevelInfo, resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}

			j, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			err = json.Unmarshal(j, &person)
			if err != nil {
				logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			go func() {
				_, err := h.coll.InsertOne(context.Background(), person)
				if err != nil {
					logger.Log(context.Background(), slog.LevelInfo, fmt.Sprintf("%s", err))
					return
				}

				err = h.redisClient.HSet(context.Background(), redisReq, person).Err()
				if err != nil {
					logger.Log(context.Background(), slog.LevelInfo, fmt.Sprintf("%s", err))
				}
			}()
		} else if err != nil {
			logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}

	p, err := json.Marshal(person)
	if err != nil {
		logger.Log(req.Context(), slog.LevelInfo, fmt.Sprintf("%s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(p)

}
