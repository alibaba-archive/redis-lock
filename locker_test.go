package redislock

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	redis "gopkg.in/redis.v5"
)

const payment = "payment"

var locker *Locker
var lockerWait *Locker

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var err error
	locker, err = NewLocker([]*redis.Client{client}, Options{})
	if err != nil {
		log.Println(err)
	}
	lockerWait, err = NewLocker([]*redis.Client{client}, Options{
		WaitTimeout: 2 * time.Second,
	})
	if err != nil {
		log.Println(err)
	}
}
func TestWrongOpts(t *testing.T) {
	assert := assert.New(t)
	locker, err := NewLocker([]*redis.Client{}, Options{})
	assert.NotNil(err)
	assert.Nil(locker)

	locker, err = NewLocker([]*redis.Client{&redis.Client{}, &redis.Client{}}, Options{})
	assert.NotNil(err)
	assert.Nil(locker)
}
func TestLocker(t *testing.T) {
	assert := assert.New(t)
	lock, err := locker.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	nillock, err := locker.Lock(payment)
	assert.NotNil(err)
	assert.Nil(nillock)

	lock.Release()

	lock, err = locker.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	lock.Release()

	lock, err = locker.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	nillock, err = locker.Lock(payment)
	assert.NotNil(err)
	assert.Nil(nillock)

	for i := 0; i < 100; i++ {
		go func() {
			lock.Release()
		}()
	}
	time.Sleep(100 * time.Millisecond)
}
func TestLockerWait(t *testing.T) {

	assert := assert.New(t)
	lock, err := lockerWait.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	// wait 2 second
	nillock, err := lockerWait.Lock(payment)
	assert.NotNil(err)
	assert.Nil(nillock)

	go func() {
		time.Sleep(time.Second)
		lock.Release()
	}()
	// successfull wait
	newlock, err := lockerWait.Lock(payment)
	assert.Nil(err)
	assert.NotNil(newlock)
}
