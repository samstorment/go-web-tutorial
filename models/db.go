package models

import (
	// redis database
	"github.com/go-redis/redis"
)

// global object -> databse client
var client *redis.Client

func Init() {
	// instantiate database client object
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", 
	})
}