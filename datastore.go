package hc

import (
    "time"
    "github.com/garyburd/redigo/redis"
    "github.com/satori/go.uuid"
)

func newPool() *redis.Pool {
    return &redis.Pool{
        MaxIdle: 3,
        IdleTimeout: 240 * time.Second,
        Dial: func () (redis.Conn, error) {
            return redis.Dial("tcp", ":6379")
        },
    }
}

var pool = newPool()

func checksByTeam() []string {
    conn := pool.Get()
    defer conn.Close()

    codes, err := redis.Strings(conn.Do("ZRANGE", "checks:admin", 0, 100))
    if err != nil {
        panic(err)
    }

    return codes
}

func dsAddCheck() {
    conn := pool.Get()
    defer conn.Close()

    now := time.Now().Unix()
    code := uuid.NewV4().String()
    conn.Do("ZADD", "checks:admin", now, code)
    conn.Do("HSET", "check:" + code, "user", "admin")
}