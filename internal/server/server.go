package server

import (
	"fmt"
	"net/http"
)

// function to ping the Redis cache and return a response
func pingHandler(w http.ResponseWriter, r *http.Request) {

}

// function to wrap a readiness probe around - will not return 200 unless Redis is available
func readyzHandler(w http.ResponseWriter, r *http.Request) {

}

func Run() {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", "5678"),
	}

	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
