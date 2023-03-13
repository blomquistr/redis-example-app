package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	redisCache "github.com/blomquistr/go-redis-example/v2/internal/cache"
	"github.com/redis/go-redis/v9"
	"k8s.io/klog"
)

var (
	ctx    context.Context = context.TODO()
	rdb    *redisCache.Database
	config IConfig
)

// function to ping the Redis cache and return a response
func pingHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Handling a ping...")
	result, err := rdb.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	klog.Info(fmt.Sprintf("Received response [%s] from Redis", result))
	w.Write([]byte(result))
}

// function to wrap a readiness probe around - will not return 200 unless Redis is available
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Handling a readiness probe...")
	result, err := rdb.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	klog.Info(fmt.Sprintf("Received response [%s] from Redis", result))
	w.Write([]byte(result))
}

// function to dump some debugging information - keep tacking on more debug info later
func debugHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Dumping debug information...")
	w.Write([]byte(fmt.Sprintf("Configuration:\n==========\n[%+v]\n", config)))
	w.Write([]byte(fmt.Sprintf("Variables:\n==========\ncontext: [%+v]\n==========\nrdb: [%+v]\n==========\n", ctx, rdb)))
}

// a wrapper function to validate we're getting the right method
// from the caller; takes two parameters, a list of supported
// methods and the method from the caller. If the method from the
// caller is in the list of supported methods, it returns nil.
// Otherwise, this method returns an error listing the method
// given and the supported methods of the calling function.
func checkSupportedMethod(methods []string, method string) error {
	for _, v := range methods {
		if v == method {
			return nil
		}
	}

	return errors.New(
		fmt.Sprintf("Invalid request method [%s], supported methods are [%s]", method, methods),
	)
}

// a struct representing a request to write a value
// to our Redis cache. Yes, it's bullshit, but it's a
// good example of how a request in a Go web server
// would be handled
type WriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// make a Redis database entry
func makeWorkHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Making some work in Redis...")

	// check to make sure we have the right request method, and
	// if not return that information to the caller to re-submit
	// their request
	klog.Info("Checking request type...")
	methods := []string{"PUT", "POST"}
	err := checkSupportedMethod(methods, r.Method)
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
		msg := fmt.Sprintf("Invalid request method [%s], supported methods are [%s]", r.Method, "PUT, POST")
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// we're going to start by constructing our message request;
	// notice how we're setting the TTL but leaving the other
	// values blank. We will accept the user omitting the TTL
	// value, but they must provide a key and a message for our
	// silly little make-work exercise
	klog.Info("Creating a new WriteRequest struct...")
	m := WriteRequest{
		TTL: config.getDefaultTTL(),
	}

	klog.Info("Decoding the JSON body...")
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
	klog.Info(fmt.Sprintf("Writing request [%v] value to Redis...", m))
	resp, err := rdb.Set(m.Key, m.Value, m.TTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write the response back to the caller; this will provide a status code
	klog.Info("Responding to the caller...")
	w.Write([]byte(resp))
}

type ReadRequest struct {
	Key string `json:"key"`
}

type ReadResult struct {
	Value string `json:"value"`
}

// read an entry from the database
func readCacheHandler(w http.ResponseWriter, r *http.Request) {
	klog.Info("Reading something from the Redis cache...")

	klog.Info("Checking request type...")
	methods := []string{"GET"}
	err := checkSupportedMethod(methods, r.Method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	switch r.Method {
	case "GET":
		klog.Info("Processing GET request to retrieve a cache entry")
	default:
		msg := fmt.Sprintf("Invalid request method [%s], supported methods are [%s]", r.Method, "GET")
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// we're going to start by constructing our message request;
	// notice how we're setting the TTL but leaving the other
	// values blank. We will accept the user omitting the TTL
	// value, but they must provide a key and a message for our
	// silly little make-work exercise
	klog.Info("Creating a new WriteRequest struct...")
	m := ReadRequest{}

	klog.Info("Decoding the JSON body...")
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

	// now we have a key, lets read it from the Redis database
	result, err := rdb.Get(m.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// finally, we want to return our value as JSON, so we're
	// going to use json.Marshal to convert it. The use of a
	// struct will let us tell the Marshal call what to map
	// the value to.
	err = encodeJSONBody(w, ReadResult{
		Value: result,
	})

	// whoops, invalid JSON, better write an error to the stream!
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	// create the new database; note how we have to create var err
	// or we will lose rdb as the global variable and have only a
	// variable scoped to our function. Fuckery, I say... :P
	var err error
	rdb, err = redisCache.NewRedisDatabase(&opts, &ctx)

	if err != nil {
		klog.Errorf("Error encountered connecting to Redis cache.")
		klog.Errorf("Configuration:\n==========\n[%+v]\n", config)
		klog.Errorf("Redis options:\n==========\n[%+v]\n", opts)
		klog.Fatal(err)
	}

	// I know we don't need to ping here because we're doing it at
	// object creation, but it still gives me comfort to know we can
	_, err = rdb.Ping()
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
