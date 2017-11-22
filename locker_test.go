package redislock

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redis "gopkg.in/redis.v5"
)

var locker *Locker
var lockerWait *Locker

func init() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	log.Println(client.Ping())
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
	l, err := NewLocker([]*redis.Client{}, Options{})
	assert.NotNil(err)
	assert.Nil(l)

	l, err = NewLocker([]*redis.Client{&redis.Client{}, &redis.Client{}}, Options{})
	assert.NotNil(err)
	assert.Nil(l)
}
func TestLocker(t *testing.T) {
	payment := "TestLocker:payment"
	assert := assert.New(t)
	require := require.New(t)
	lock, err := locker.Lock(payment)
	require.Nil(err)
	require.NotNil(lock)
	assert.Equal(100*time.Millisecond, lock.opts.WaitRetry)
	assert.Equal(10*time.Second, lock.opts.LockTimeout)
	assert.Equal(time.Duration(0), lock.opts.WaitTimeout)

	start := time.Now()
	nillock, err := locker.Lock(payment)
	assert.True(time.Since(start) < 50*time.Millisecond)
	assert.NotNil(err)
	require.Nil(nillock)

	require.NotNil(lock)
	lock.Unlock()

	lock, err = locker.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	lock.Unlock()

	lock, err = locker.Lock(payment)
	assert.Nil(err)
	assert.NotNil(lock)

	nillock, err = locker.Lock(payment)
	assert.NotNil(err)
	assert.Nil(nillock)

	for i := 0; i < 100; i++ {
		go func() {
			lock.Unlock()
		}()
	}
	time.Sleep(100 * time.Millisecond)
}
func TestLockerWait(t *testing.T) {
	payment := randomValue()

	assert := assert.New(t)
	lock, err := lockerWait.Lock(payment)
	assert.Equal(defaultRedisKeyPrefix+payment, lock.lockkey)
	assert.Nil(err)
	assert.NotNil(lock)

	// wait 2 second
	nillock, err := lockerWait.Lock(payment)
	assert.Equal(ErrGetLockFailed, err)
	assert.Nil(nillock)

	go func() {
		time.Sleep(time.Second)
		if assert.NotNil(lock) {
			lock.Unlock()
		}
	}()
	// successfull wait
	newlock, err := lockerWait.Lock(payment)
	assert.Nil(err)
	assert.NotNil(newlock)
}

func TestWrongRedis(t *testing.T) {
	assert := assert.New(t)

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:9999",
	})
	locker, err := NewLocker([]*redis.Client{client}, Options{})
	assert.Nil(err)
	_, err = locker.Lock("x")
	assert.Equal(ErrGetLockFailed, err)
}
