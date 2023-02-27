package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"k8s.io/klog"
)

var (
	ctx    context.Context = context.TODO()
	rdb    redis.Client
	config IConfig
)

// function to ping the Redis cache and return a response
func pingHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Handling a ping...")
	w.Write([]byte("pong"))
}

// function to wrap a readiness probe around - will not return 200 unless Redis is available
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Handling a readiness probe...")
	w.Write([]byte("ok"))
}

// function to dump some debugging information - keep tacking on more debug info later
func debugHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Dumping debug information...")
	w.Write([]byte(fmt.Sprintf("Configuration:\n==========\n[%+v]\n", config)))
}

// function to parse the TTL a user provides for their Redis cache entry and take the default value otherwise
func parseRequestTTL(r *http.Request) (time.Duration, error) {
	requestTTL := r.FormValue("ttl")
	var ttl time.Duration
	var err error
	if requestTTL == "" {
		ttl = time.Duration(config.getDefaultTTL())
	} else {
		intTTL, err := strconv.Atoi(requestTTL)
		if err != nil {
			return ttl, err
		}
		ttl = time.Duration(intTTL)
	}

	return ttl, err
}

// make a Redis database entry
func makeWorkHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Making some work in Redis...")
	w.Write([]byte("Not implemented"))
}

// read an entry from the database
func readCacheHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Reading something from the Redis cache...")
	w.Write([]byte("Not implemented"))
}

// this is the place we actually start the server.
func Run() {
	// first thing's first, lets load our configuration using the config.go
	// interface we defined for our server.
	config = newConfig()

	// next, we need to define some endpoints for the server to handle
	// in this we're binding a specific endpoint (the string parameter)
	// to a specific handler function. You can either define the function
	// inline, or create a separate one. Because I feel it creates a
	// more readable piece of code, I've elected to define separate
	// functions for each endpoint handler.
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/healthz", readyzHandler)
	http.HandleFunc("/debug", debugHandler)

	// these two handlers are going to do some BS work against our Redis
	// implementations. Sending a request to write-redis will
	http.HandleFunc("/write-redis", makeWorkHandler)
	http.HandleFunc("/read-redis", readCacheHandler)

	// next, lets start our Redis connection!
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.getRedisAddress(), config.getRedisPort()),
		Password: config.getRedisPassword(),
		DB:       config.getRedisDB(),
	})

	// validate that Redis is up
	pong, err := rdb.Ping(ctx).Result()

	if err != nil {
		klog.Errorf("Error encountered connecting to Redis cache.")
		klog.Errorf("Configuration:\n==========\n[%+v]\n", config)
		klog.Fatal(err)
	}

	klog.Infof("Connected to Redis database and received pong [%s] when testing the connection", pong)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.getPort()),
	}

	err = server.ListenAndServe()

	if err != nil {
		klog.Fatal(err)
	}
}
