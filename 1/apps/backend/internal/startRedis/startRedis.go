package startRedis

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wb-go/wbf/redis"
)

func StartRedis() *redis.Client {
	port := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), port)
	password := os.Getenv("REDIS_PASSWORD")
	database, err := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	if err != nil {
		panic(err)
	}

	client := redis.New(addr, password, database)

	return client
}
