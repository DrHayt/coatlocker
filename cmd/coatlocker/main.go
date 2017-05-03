package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/drhayt/coatlocker/pkg/fshandler"
	"github.com/drhayt/coatlocker/pkg/jwtclient"
	hndl "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	var (
		certURL       = flag.String("certurl", "https://authentication.sgtec.io/Certificate", "The directory to use as the base of file uploads/downloads")
		baseDirectory = flag.String("basedir", "/tmp", "The directory to use as the base of file uploads/downloads")
		listenPort    = flag.String("port", "8443", "The port to listen on")
		listenAddress = flag.String("address", "0.0.0.0", "The address to listen on")
		insecure      = flag.Bool("insecure", false, "Do not validate https certificates")
		certPath      = flag.String("certpath", "server.crt", "The path to the certificate")
		keyPath       = flag.String("pathpath", "server.key", "The path to the key")
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

	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(*listenAddress, *listenPort), *certPath, *keyPath, h))

}
