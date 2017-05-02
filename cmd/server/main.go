package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/drhayt/coatlocker/pkg/fshandler"
	hndl "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	var (
		baseDirectory = flag.String("basedir", "/tmp", "The directory to use as the base of file uploads/downloads")
		listenPort    = flag.Int("port", 8443, "The port to listen on")
		listenAddress = flag.String("address", "127.0.0.1", "The address to listen on")
	)
	flag.Parse()

	// Get a copy of the server struct to work with
	server := fshandler.Server{BaseDirectory: *baseDirectory}

	// Validate our server config.
	err := server.Validate()
	if err != nil {
		panic("Invalid base directory")
	}

	// Logging middleware.
	router := mux.NewRouter()
	h := hndl.LoggingHandler(os.Stdout, router)

	//Users
	router.HandleFunc("/health", server.HealthEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.GetEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.PutEndpoint).Methods("PUT")
	router.PathPrefix("/").HandlerFunc(server.DeleteEndpoint).Methods("DELETE")

	// This sets up the application to serve.
	go http.ListenAndServe(":8887", nil) // For net/trace
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *listenAddress, *listenPort), h))

	// This sets up the tracing for requests.
	// log.Fatal(http.ListenAndServe(":8887", nil)) // For net/trace

	// We should setup an additional GRPC endpoint.

}
