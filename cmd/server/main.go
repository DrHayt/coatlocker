package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	hndl "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/safeguardproperties/coatlocker/pkg/fshandler"
	"github.com/safeguardproperties/coatlocker/pkg/jwtclient"
)

func main() {

	var (
		certURL       = flag.String("certurl", "https://authentication.sgtec.io/Certificate", "The directory to use as the base of file uploads/downloads")
		baseDirectory = flag.String("basedir", "/tmp", "The directory to use as the base of file uploads/downloads")
		listenPort    = flag.Int("port", 8443, "The port to listen on")
		listenAddress = flag.String("address", "127.0.0.1", "The address to listen on")
		insecure      = flag.Bool("insecure", false, "Do not validate https certificates")
	)
	flag.Parse()

	var server ICoatHandler

	// Get a copy of the server struct to work with
	server = fshandler.Server{BaseDirectory: *baseDirectory}

	// Validate our server config.
	err := server.Validate()
	if err != nil {
		panic("Invalid base directory")
	}

	// Get a closure to use with the jwt stuff.
	keyFunc, err := jwtclient.KeyFuncClosure(*insecure, *certURL)
	if err != nil {
		panic("Unable to create key validation function")
	}

	// middleware order from innermost to outermost.
	router := mux.NewRouter()
	// Setup jwt middleware.
	options := jwtmiddleware.Options{
		SigningMethod:       jwt.SigningMethodRS256,
		ValidationKeyGetter: keyFunc,
	}
	jwthandler := jwtmiddleware.New(options)

	// Wrap the router with the jwt handler.
	h := jwthandler.Handler(router)
	// Wrap a logger around everything.
	h = hndl.LoggingHandler(os.Stdout, h)

	// CoatLocker
	router.HandleFunc("/health", server.HealthEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.GetEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.PutEndpoint).Methods("PUT")
	router.PathPrefix("/").HandlerFunc(server.DeleteEndpoint).Methods("DELETE")

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *listenAddress, *listenPort), h))

}
