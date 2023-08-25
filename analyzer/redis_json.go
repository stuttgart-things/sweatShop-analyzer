package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	redigoredis "github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
	goredis "github.com/redis/go-redis/v9"
)

// ErrJSONMissWithGoRedisClient has ideally the type RedisError of "github.com/redis/go-redis/v9/internal/proto"
var ErrJSONMissWithGoRedisClient = "redis: nil"
var ErrJSONMissWithRedigoConn = errors.New("redigo: nil returned")

type AnalyzerResultValue struct {
	Repo    *Repository
	Commit  string
	Results []*TechAndPath
}

type AnalyzerJSONHandler struct {
	client  *goredis.Client
	conn    *redigoredis.Conn
	handler *rejson.Handler
}

// attention: go-rejson/v4@v4.1.0 does not support redis/go-redis/v9 (but redis/go-redis/v8)
func NewAnalyzerJSONHandlerWithGoRedisClient(client *goredis.Client) *AnalyzerJSONHandler {

	// Create a new ReJSON instance
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClientWithContext(context.Background(), client)

	return &AnalyzerJSONHandler{
		client:  client,
		handler: rh,
	}
}

func NewAnalyzerJSONHandlerWithRedigoConn(rs string) *AnalyzerJSONHandler {

	// Connect to Redis server
	conn, err := redigoredis.Dial("tcp", rs)
	if err != nil {
		log.Fatalf("Could not connect to Redis server: %v\n", err)
	}

	// Create a new ReJSON instance
	rh := rejson.NewReJSONHandler()

	rh.SetRedigoClient(conn)

	return &AnalyzerJSONHandler{
		conn:    &conn,
		handler: rh,
	}
}

func analyzerResultKey(repoURL string) string {
	return fmt.Sprintf("analyzerresult|%s", repoURL)
}

func (h *AnalyzerJSONHandler) SetAnalyzerResult(repo *Repository, commitId string, res []*TechAndPath) error {
	item := &AnalyzerResultValue{repo, commitId, res}
	return h.SetItem(analyzerResultKey(repo.Url), item, false)
}

func (h *AnalyzerJSONHandler) GetAnalyzerResult(repoURL string) (*AnalyzerResultValue, error) {
	item := &AnalyzerResultValue{}
	return item, h.GetItem(analyzerResultKey(repoURL), item)
}

func (h *AnalyzerJSONHandler) SetItem(key string, item interface{}, delete bool) error {
	if delete {
		return h.Delete(key)
	} else {
		if item == nil {
			return fmt.Errorf("cannot set item to nil for key %s", key)
		}

		// Add a JSON set command
		res, err := h.handler.JSONSet(key, ".", item)
		if err != nil {
			return fmt.Errorf("could not set JSON item: %v", err)
		}

		if res.(string) != "OK" {
			return fmt.Errorf("could not set JSON item: %v", res)
		}
	}

	return nil
}

func (h *AnalyzerJSONHandler) GetItem(key string, item interface{}) error {
	if item == nil {
		return fmt.Errorf("cannot get item into a nil for key %s", key)
	}

	itemByte, err := redigoredis.Bytes(h.handler.JSONGet(key, "."))
	if err != nil {
		return err
	}

	err = json.Unmarshal(itemByte, item)
	if err != nil {
		return fmt.Errorf("failed to JSON Unmarshal: %v", err)
	}

	return nil
}

func (h *AnalyzerJSONHandler) Delete(key string) error {
	res, err := h.handler.JSONDel(key, ".")
	if err != nil {
		return fmt.Errorf("could not JSONDel: %v", err)
	}
	if res.(int64) != 1 {
		return fmt.Errorf("could not JSONDel: %v", res)
	}
	return nil
}
