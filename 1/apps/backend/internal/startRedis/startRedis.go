package startRedis

import (
	"os"
	"strconv"

	"github.com/wb-go/wbf/redis"
)

func StartRedis() *redis.Client {
	addr := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	database, err := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	if err != nil {
		panic(err)
	}

	client := redis.New(addr, password, database)

	return client
}
