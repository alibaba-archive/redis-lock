# redis-lock
[![Build Status](https://img.shields.io/travis/mushroomsir/redis-lock.svg?style=flat-square)](https://travis-ci.org/mushroomsir/redis-lock)
[![Coverage Status](http://img.shields.io/coveralls/mushroomsir/redis-lock.svg?style=flat-square)](https://coveralls.io/github/mushroomsir/redis-lock?branch=master)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://github.com/mushroomsir/redis-lock/blob/master/LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/mushroomsir/redis-lock)

## Installation

```sh
go get github.com/mushroomsir/redis-lock
```

## Usage
```go
package main

import (
	"fmt"
	"log"

	"github.com/mushroomsir/redis-lock"
	redis "gopkg.in/redis.v5"
)

const payment = "payment"

var locker *redislock.Locker

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var err error
	locker, err = redislock.NewLocker([]*redis.Client{client}, redislock.Options{})
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
	defer lock.Release()

	fmt.Println("have a lock")

	lock, err = locker.Lock(payment)
	if err != nil {
		fmt.Println(err)
	}
	// output:
	// have a lock
	// get lock failed
}
```

## Licenses

All source code is licensed under the [MIT License](https://github.com/mushroomsir/redis-lock/blob/master/LICENSE).
