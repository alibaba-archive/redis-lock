package main

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
	redislock "github.com/teambition/redis-lock"
)

const payment = "payment"

var locker *redislock.Locker

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "L7PxpiPPHwFApUhv",
	})
	var err error
	locker, err = redislock.NewLocker([]redis.Cmdable{client}, redislock.Options{})
	if err != nil {
		log.Println(err)
	}
}

func main() {

	lock, err := locker.Lock(payment)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lock.Unlock()

	fmt.Println("have a lock")

	lock, err = locker.Lock(payment)
	if err != nil {
		fmt.Println(err)
	}
	// output:
	// have a lock
	// get lock failed
}
