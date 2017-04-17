package main

import (
    "fmt"
    "io"
    "net/http"
    "strconv"
    "time"
    "github.com/go-redis/redis"
)

var client *redis.Client

func handlePing(code string, arrived int64, ip string, ua string) {
    key := "check:" + code
    logkey := key + ":log"

    // Get timeout and grace
    fields := client.HMGet(key, "status", "timeout", "grace").Val()
    status := fields[0]

    // Set last ping and n_pings
    client.HSet(key, "last_ping", arrived)
    nPings := client.HIncrBy(key, "n_pings", 1).Val()
    if status == nil || status == "paused" {
        client.HSet(key, "status", "up")
    }

    // Add log entry
    client.LPush(logkey, fmt.Sprintf("%d|%d|%s|%s", nPings, arrived, ip, ua))
    if (nPings % 10 == 0) {
        client.LTrim(logkey, 0, 99)
    }

    // Schedule next alert
    alertTime := arrived
    if status != "down" {
        timeout, _ := strconv.ParseInt(fields[1].(string), 10, 64)
        grace, _ := strconv.ParseInt(fields[2].(string), 10, 64)
        alertTime += timeout + grace
    }

    alert := redis.Z{float64(alertTime), code}
    client.ZAdd("alerts", alert)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Path[1:]
    if (len(code) != 36) {
        http.Error(w, "", 400)
        return
    }

    ip := r.Header.Get("X-Forwarded-For")
    ua := r.Header.Get("User-Agent")
    if len(ua) > 50 {
        ua = ua[0:50]
    }

    fmt.Println("Ping " + code + " " + ip)

    handlePing(code, time.Now().Unix(), ip, ua)
    io.WriteString(w, "OK")
}

func main() {
    client = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB: 0,
    })

    pong, err := client.Ping().Result()
    fmt.Println(pong, err)

    http.HandleFunc("/", httpHandler)
    http.ListenAndServe("0.0.0.0:8000", nil)
}