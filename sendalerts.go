package main

import (
    "fmt"
    "time"
    "github.com/garyburd/redigo/redis"
)

var conn redis.Conn

func process() {
    now := time.Now().Unix()

    codes, err := redis.Strings(conn.Do("ZRANGEBYSCORE", "alerts", "0", now))
    if err != nil {
        panic(err)
    }

    for _, code := range codes {
        lock := "sendalerts:" + code + ":lock"

        // Lock
        locked, err := redis.String(conn.Do("SET", lock, "EX 60", "NX"))
        if err != nil {
            panic(err)
        }

        if locked != "OK" {
            fmt.Println(code + " is already locked")
            continue
        }

        fmt.Println("Processing " + code)
        // FIXME actually decide what to do here

        // Remove it from task pool
        conn.Do("ZREM", "alerts", code)

        // Unlock
        conn.Do("DEL", lock)
    }
}

func main() {
    var err error
    conn, err = redis.Dial("tcp", ":6379")
    if err != nil {
        panic(err)
    }

    defer conn.Close()

    process()
}