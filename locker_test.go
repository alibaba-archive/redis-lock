package redislock

import (
	"log"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var locker *Locker
var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	log.Println(redisClient.Ping())
	var err error
	locker, err = NewLocker([]redis.Cmdable{redisClient}, Options{})
	if err != nil {
		log.Println(err)
	}

}
func TestWrongOpts(t *testing.T) {
	assert := assert.New(t)
	l, err := NewLocker([]redis.Cmdable{}, Options{})
	assert.NotNil(err)
	assert.Nil(l)

	l, err = NewLocker([]redis.Cmdable{&redis.Client{}, &redis.Client{}}, Options{})
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
	assert := assert.New(t)
	require := require.New(t)
	lockerWait, err := NewLocker([]redis.Cmdable{redisClient}, Options{
		KeyPrefix:   "test:",
		LockTimeout: 300 * time.Millisecond,
		WaitTimeout: 200 * time.Millisecond,
	})
	require.Nil(err)
	require.NotNil(lockerWait)

	payment := randomValue()
	// successfull lock
	lock, err := lockerWait.Lock(payment)
	assert.Equal(lockerWait.opts.KeyPrefix+payment, lock.lockkey)
	assert.Nil(err)
	assert.NotNil(lock)

	// failed
	nillock, err := lockerWait.Lock(payment)
	assert.Equal(ErrGetLockFailed, err)
	assert.Nil(nillock)

	go func() {
		time.Sleep(10 * time.Millisecond)
		require.NotNil(lock)
		lock.Unlock()
	}()
	// successfull wait
	newlock, err := lockerWait.Lock(payment)
	assert.Nil(err)
	assert.NotNil(newlock)

	// failed
	nillock, err = lockerWait.Lock(payment)
	assert.Equal(ErrGetLockFailed, err)
	assert.Nil(nillock)

	// the lock was expired, can lock
	time.Sleep(100 * time.Millisecond)
	newlock, err = lockerWait.Lock(payment)
	assert.Nil(err)
	assert.NotNil(newlock)
}

func TestWrongRedis(t *testing.T) {
	assert := assert.New(t)

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:9999",
	})
	locker, err := NewLocker([]redis.Cmdable{client}, Options{})
	assert.Nil(err)
	l, err := locker.Lock("x")
	assert.Nil(l)
	assert.Equal(ErrGetLockFailed, err)
}
