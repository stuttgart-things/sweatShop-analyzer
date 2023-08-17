/*
Copyright © 2023 PATRICK HERMANN patrick.hermann@sva.de
*/

package stream

import (
	"fmt"
	"net/url"
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
	repo := buildValidRepository(msg.Values)
	if repo == nil {
		log.Error("INVALID INPUT RECEIVED")
		return nil
	}

	repo.GetMatchingFiles()

	return nil
}

func buildValidRepository(values map[string]interface{}) *analyzer.Repository {

	if len(values) == 0 {
		log.Error("NO VALUES RECEIVED")
		return nil
	}

	if values["url"] == nil {
		log.Error("NO URL RECEIVED")
		return nil
	}

	_, err := url.ParseRequestURI(values["url"].(string))
	if err != nil {
		log.Errorf("INVALID URL RECEIVED: %s", values["url"].(string))
		return nil
	}

	if values["revision"] == nil {
		log.Error("NO REVISION RECEIVED")
		return nil
	}

	// try to construct repository using the received values
	r := &analyzer.Repository{
		Url:      values["url"].(string),
		Revision: values["revision"].(string),
	}
	if values["name"] != nil {
		r.Name = values["name"].(string)
	}
	if values["username"] != nil {
		r.Username = values["username"].(string)
	}
	if values["password"] != nil {
		r.Password = values["password"].(string)
	}
	if values["insecure"] != nil {
		r.Insecure = sthingsBase.ConvertStringToBoolean(values["insecure"].(string))
	}

	// try to connect to the repository
	err = r.ConnectRepository()
	if err != nil {
		log.Errorf("COULD NOT CONNECT TO REPOSITORY: %s", err.Error())
		return nil
	}

	// TODO: check if the revision exists

	return r
}
