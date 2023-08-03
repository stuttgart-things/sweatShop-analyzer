/*
Copyright © 2023 PATRICK HERMANN patrick.hermann@sva.de
*/

package stream

import (
	"fmt"
	"os"
	"time"

	"github.com/stuttgart-things/sweatShop-analyzer/analyzer"

	"github.com/redis/go-redis/v9"
	"github.com/stuttgart-things/redisqueue"
	sthingsBase "github.com/stuttgart-things/sthingsBase"
)

var (
	redisServer   = os.Getenv("REDIS_SERVER")
	redisPort     = os.Getenv("REDIS_PORT")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	redisStream   = os.Getenv("REDIS_STREAM")
	log           = sthingsBase.StdOutFileLogger(logfilePath, "2006-01-02 15:04:05", 50, 3, 28)
	logfilePath   = "/tmp/sweatShop-analyzer.log"
)

func PollRedisStreams() {

	c, err := redisqueue.NewConsumerWithOptions(&redisqueue.ConsumerOptions{
		VisibilityTimeout: 60 * time.Second,
		BlockingTimeout:   5 * time.Second,
		ReclaimInterval:   1 * time.Second,
		BufferSize:        100,
		Concurrency:       10,
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     redisServer + ":" + redisPort,
			Password: redisPassword,
			DB:       0,
		}),
	})

	if err != nil {
		panic(err)
	}

	c.Register(redisStream, processStreams)

	go func() {
		for err := range c.Errors {
			fmt.Printf("err: %+v\n", err)
		}
	}()

	if redisStream == "" {
		log.Error("NO STREAM DEFINED - VARIABLE REDIS_STREAM SEEMS TO BE EMPTY")
	}

	log.Info("START POLLING STREAM ", redisStream+" ON "+redisServer+":"+redisPort)

	c.Run()

	log.Warn("POLLING STOPPED")

}

func processStreams(msg *redisqueue.Message) error {

	// ADD VALUE VALIDATION HERE

	repo := analyzer.Repository{
		msg.Values["name"].(string),
		msg.Values["url"].(string),
		msg.Values["revision"].(string),
		msg.Values["username"].(string),
		msg.Values["password"].(string),
		sthingsBase.ConvertStringToBoolean(msg.Values["insecure"].(string)),
	}

	analyzer.GetMatchingFiles(repo)

	return nil
}
