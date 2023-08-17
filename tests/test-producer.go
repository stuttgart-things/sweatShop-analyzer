package main

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/stuttgart-things/redisqueue"
)

var (
	redisServer   = os.Getenv("REDIS_SERVER")
	redisPort     = os.Getenv("REDIS_PORT")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	redisStream   = os.Getenv("REDIS_STREAM")

	tests = []test{
		{testValues: ValuesRepo, testKey: "unset"},
	}

	ValuesRepo = map[string]interface{}{
		"name":                    "stuttgart-things",
		"url":                     "https://github.com/stuttgart-things/stuttgart-things.git",
		"revision":                "main",
		"username":                "",
		"password":                "",
		"insecure":                "false",
		"force_complete_analysis": "false",
	}
)

type test struct {
	testValues map[string]interface{}
	testKey    string
}

func main() {

	fmt.Println("REDIS-SERVER: " + redisServer + ":" + redisPort)
	fmt.Println("REDIS-STREAM: " + redisStream)

	// CREATE RESOURCES IN REDIS
	p, err := redisqueue.NewProducerWithOptions(&redisqueue.ProducerOptions{
		ApproximateMaxLength: true,
		RedisClient: redis.NewClient(&redis.Options{
			Addr:     redisServer + ":" + redisPort,
			Password: redisPassword,
			DB:       0,
		}),
	})

	if err != nil {
		panic(err)
	}

	// CREATE RESOURCES IN REDIS
	for _, tc := range tests {

		fmt.Println("\nTEST-DATA:")
		for key, data := range tc.testValues {
			fmt.Println(key, ":", data)
		}

		err2 := p.Enqueue(&redisqueue.Message{
			Stream: redisStream,
			Values: tc.testValues,
		})

		if err2 != nil {
			panic(err)
		}

		fmt.Println("\nTEST DATA WRITTEN TO REDIS STREAM", redisStream)

	}

}
