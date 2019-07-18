package common

import (
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var defaultRedis *Redis
var defaultRedisOnce sync.Once

type Redis struct {
	pool *redis.Pool
}

func GetRedis() *Redis {
	defaultRedisOnce.Do(func() {
		defaultRedis = newRedis(":6379")
	})
	return defaultRedis
}

func newRedis(server string) *Redis {
	return &Redis{
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (c redis.Conn, err error) {
				c, err = redis.Dial("tcp", server)
				return
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (r *Redis) Get(key string) (rst string, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	rst, err = redis.String(conn.Do("GET", key))
	return
}

func (r *Redis) Set(key, value string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()
	_, err = conn.Do("SET", key, value)
	return
}

func (r *Redis) Del(key string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", key)
	return
}

func (r *Redis) HExists(key, field string) (rst bool, err error) {
	conn := r.pool.Get()
	defer conn.Close()
	rst, err = redis.Bool(conn.Do("HExists", key, field))
	return
}

func (r *Redis) HGet(key, field string) (rst string, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	rst, err = redis.String(conn.Do("HGET", key, field))
	return
}

func (r *Redis) HMGet(args ...interface{}) (rst map[string]string, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	rst, err = redis.StringMap(conn.Do("HMGET", args...))
	return
}

func (r *Redis) HGetAll(key string) (rst map[string]string, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	rst, err = redis.StringMap(conn.Do("HGETALL", key))
	return
}

func (r *Redis) HSet(key, field, value string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()

	_, err = conn.Do("HSET", key, field, value)
	return
}

func (r *Redis) HMSet(args ...interface{}) (err error) {
	conn := r.pool.Get()
	defer conn.Close()

	_, err = conn.Do("HMSET", args...)
	return
}
