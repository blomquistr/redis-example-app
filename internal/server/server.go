package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	redisCache "github.com/blomquistr/go-redis-example/v2/internal/cache"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog"
)

var (
	ctx    context.Context = context.TODO()
	rdb    *redisCache.Database
	config IConfig
)

type ReadRequest struct {
	Key string
}

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

// a struct representing a request to write a value
// to our Redis cache. Yes, it's bullshit, but it's a
// good example of how a request in a Go web server
// would be handled
type Message struct {
	Key   string
	Value string
	TTL   int
}

func checkSupportedMethod(methods []string, method string) error {
	for _, v := range methods {
		if v == method {
			return nil
		}
	}

	return errors.New(
		fmt.Sprintf("Invalid request method [%s], supported methods include [%s]", method, methods),
	)
}

// make a Redis database entry
func makeWorkHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Making some work in Redis...")

	// check the supported method type and, if it is not supported, return
	// a MethodNotAllowed status to the caller
	supportedMethods := []string{"POST", "PUT"}
	err := checkSupportedMethod(supportedMethods, r.Method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// Lets make sure we have the right type of request - we only
	// want to handle POST or PUT requests.
	switch r.Method {
	case "POST":
		klog.Info("Processing POST request for new cache entry")
	case "PUT":
		klog.Info("Processing PUT request to update existing cache entry")
	default:
		msg := fmt.Sprintf("Invalid request method [%s], supported methods include [%s]", r.Method, supportedMethods)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// we're going to start by constructing our message request;
	// notice how we're setting the TTL but leaving the other
	// values blank. We will accept the user omitting the TTL
	// value, but they must provide a key and a message for our
	// silly little amke-work exercise
	m := Message{
		TTL: config.getDefaultTTL(),
	}

	err = decodeJSONBody(w, r, &m)
	// with handling of the decoding wrapped in a separate method, we can deal with
	// the errors that handler bubbles up in a more condensed way in our request
	// handler method.
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			klog.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// do something here to write to Redis
	rdb.Client.Set(ctx, m.Key, m.Value, time.Duration(int64(m.TTL)))
}

// read an entry from the database
func readCacheHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Reading something from the Redis cache...")

	switch r.Method {
	case "GET":
		klog.Info("Processing GET request to retrieve a cache entry")
	default:
		msg := fmt.Sprintf("Invalid request method [%s], supported methods include PUT and POST", r.Method)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

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
	opts := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.getRedisAddress(), config.getRedisPort()),
		Password: config.getRedisPassword(),
		DB:       config.getRedisDB(),
	}
	rdb, err := redisCache.NewRedisDatabase(&opts)

	if err != nil {
		klog.Errorf("Error encountered connecting to Redis cache.")
		klog.Errorf("Configuration:\n==========\n[%+v]\n", config)
		klog.Errorf("Redis options:\n==========\n[%+v]\n", opts)
		klog.Fatal(err)
	}

	_, err = rdb.Client.Ping(ctx).Result()
	if err != nil {
		klog.Errorf("Error pinging Redis cache")
		klog.Errorf("Configuration:\n==========\n[%+v]\n", config)
		klog.Errorf("Redis options:\n==========\n[%+v]\n", opts)
		klog.Fatal(err)
	} else {
		klog.Infof("Connected to Redis database and received pong when testing the connection")
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.getPort()),
	}

	err = server.ListenAndServe()

	if err != nil {
		klog.Fatal(err)
	}
}
