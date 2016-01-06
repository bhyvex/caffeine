package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

const HIPACHE_PREFIX = "frontend:"

var (
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", hipacheRedisAddr())
		if err != nil {
			return nil, err
		}
		return c, err
	}, hipacheRedisMaxConn())
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn := redisPool.Get()
		defer conn.Close()
		host := r.Host
		restoreRoute(host, conn)
		startApp(host)
		time.Sleep(10 * time.Second)
		proxy := createProxy(r)
		proxy.ServeHTTP(w, r)
	})

	http.ListenAndServe("0.0.0.0:8888", nil)
}

func restoreRoute(host string, conn redis.Conn) {
	name := HIPACHE_PREFIX + host
	log.Printf("Deleting %s\n", name)
	_, err := conn.Do("LTRIM", name, 0, 0)
	if err != nil {
		log.Printf("Err: %s\n", err)
	}
}

func hipacheRedisAddr() string {
	host := getConfig("HIPACHE_REDIS_HOST")
	port := getConfig("HIPACHE_REDIS_PORT")

	return fmt.Sprintf("%s:%s", host, port)
}

func hipacheRedisMaxConn() int {
	maxConn, _ := strconv.Atoi(getConfig("HIPACHE_REDIS_MAX_CONN"))
	return maxConn
}

func getConfig(key string) string {
	defaultValues := map[string]string{
		"HIPACHE_REDIS_HOST":     "localhost",
		"HIPACHE_REDIS_PORT":     "6379",
		"HIPACHE_REDIS_MAX_CONN": "10",
		"TSURU_HOST":             "http://localhost",
		"TSURU_APP_PROXY":        "",
		"TSURU_TOKEN":            "",
	}

	value := os.Getenv(key)
	if value == "" {
		return defaultValues[key]
	}

	return value
}
