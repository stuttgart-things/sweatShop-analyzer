package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/nitishm/go-rejson/v4"
	goredis "github.com/redis/go-redis/v9"
)

// ErrJSONMissWithGoRedisClient has ideally the type RedisError of "github.com/redis/go-redis/v9/internal/proto"
var ErrJSONMissWithGoRedisClient = "redis: nil"
var ErrJSONMissWithRedigoConn = errors.New("redigo: nil returned")

type Redis struct {
	Server      string
	Port        int
	Password    string
	Client      *goredis.Client
	JSONHandler *rejson.Handler
}

func newRedis(server string, port int, password string) *Redis {
	return &Redis{
		Server:   server,
		Port:     port,
		Password: password,
	}
}

func NewRedisWithClient(server string, port int, password string) *Redis {
	r := newRedis(server, port, password)
	r.Client = goredis.NewClient(&goredis.Options{
		Addr:     r.GetServerPort(),
		Password: r.Password,
		DB:       0,
	})
	return r
}

func (r *Redis) GetServerPort() string {
	return fmt.Sprintf("%s:%d", r.Server, r.Port)
}

// attention: go-rejson/v4@v4.1.0 does not support redis/go-redis/v9 (but redis/go-redis/v8)
// go-rejson/master does support redis/go-redis/v9
func (r *Redis) SetJSONHandler() {
	// Create a new ReJSON instance
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClientWithContext(context.Background(), r.Client)
	r.JSONHandler = rh
}

/*
func newJSONHandlerWithRedigoConn(rs string) *rejson.Handler {

	// Connect to Redis server
	// TODO: add authentication
	conn, err := redigoredis.Dial("tcp", rs)
	if err != nil {
		log.Fatalf("Could not connect to Redis server: %v\n", err)
	}

	// Create a new ReJSON instance
	rh := rejson.NewReJSONHandler()

	rh.SetRedigoClient(conn)

	return rh
}
*/
