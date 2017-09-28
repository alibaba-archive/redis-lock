package redislock

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	redis "gopkg.in/redis.v5"
)

var (
	ErrGetLockFailed = errors.New("get lock failed")
	ErrGetNewLock    = errors.New("please use Locker to get lock")
	ErrNoRedisClient = errors.New("no redis clients")
	ErrMustBeOdd     = errors.New("the redis count must be odd")
)

const (
	defaultRedisKeyPrefix = "redislock:"
	luaRelease            = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
)

// NewLocker ...
func NewLocker(clients []*redis.Client, opts Options) (*Locker, error) {
	if len(clients) == 0 {
		return nil, ErrNoRedisClient
	}
	if len(clients)%2 == 0 {
		return nil, ErrMustBeOdd
	}
	if opts.KeyPrefix == "" {
		opts.KeyPrefix = defaultRedisKeyPrefix
	}
	if opts.WaitRetry == 0 {
		opts.WaitRetry = 100 * time.Millisecond
	}
	if opts.LockTimeout == 0 {
		opts.LockTimeout = 10 * time.Second
	}
	return &Locker{
		clients: clients,
		opts:    &opts,
	}, nil
}

// Locker ...
type Locker struct {
	clients []*redis.Client
	opts    *Options
}

// Options ...
type Options struct {
	KeyPrefix string
	// The maximum duration to lock a key, Default: 10s
	LockTimeout time.Duration
	// The maximum duration to wait to get the lock, Default: 0s, do not wait
	WaitTimeout time.Duration
	// The maximum wait retry time to get the lock again, Default: 100ms
	WaitRetry time.Duration
}

// Lock ...
type Lock struct {
	session string
	lockkey string
	clients []*redis.Client
	opts    *Options
	clock   sync.Mutex
}

// Release the lock
func (l *Lock) Release() {
	l.clock.Lock()
	defer l.clock.Unlock()
	if l.lockkey == "" {
		return
	}
	key := l.lockkey
	l.lockkey = ""
	for _, client := range l.clients {
		client.Eval(luaRelease, []string{key}, l.session)
	}
}

// Lock lock
func (l *Lock) lock() error {
	stop := time.Now().Add(l.opts.WaitTimeout)
	for {
		successful := 0
		start := time.Now()
		for _, client := range l.clients {
			if client.SetNX(l.lockkey, l.session, l.opts.LockTimeout).Val() {
				successful++
			}
		}
		deviation := l.opts.LockTimeout / 100
		validity := l.opts.LockTimeout - time.Since(start) - deviation
		if validity > 0 && successful > len(l.clients)/2 {
			return nil
		}
		if time.Now().Add(l.opts.WaitRetry).After(stop) {
			break
		}
		time.Sleep(l.opts.WaitRetry)
	}
	return ErrGetLockFailed
}

// Lock lock
func (l *Locker) Lock(key string) (*Lock, error) {
	lock := &Lock{
		session: randomValue(),
		lockkey: key,
		clients: l.clients,
		opts:    l.opts,
	}
	err := lock.lock()
	if err != nil {
		return nil, err
	}
	return lock, nil
}

func randomValue() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf)
}
