package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"bytes"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/drhayt/coatlocker/pkg/fshandler"
	"github.com/drhayt/jwtclient"
	hndl "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func main() {

	var (
		baseDirectory = flag.String("basedir", os.Getenv("COATLOCKER_BASEDIR"), "The directory to use as the base of file uploads/downloads")
		listenPort    = flag.String("port", os.Getenv("COATLOCKER_PORT"), "The port to listen on")
		listenAddress = flag.String("address", os.Getenv("COATLOCKER_ADDRESS"), "The address to listen on")
		certPath      = flag.String("certpath", os.Getenv("COATLOCKER_CERTPATH"), "The path to the certificate")
		keyPath       = flag.String("keypath", os.Getenv("COATLOCKER_KEYPATH"), "The path to the key")
		jwtCertPath   = flag.String("jwtcertpath", os.Getenv("COATLOCKER_JWTCERTPATH"), "The path to the PEM encoded JWT certificate to validate against")
	)
	flag.Parse()

	// FIXME:   Do more argument checking stuff.

	var server ICoatHandler

	// Get a copy of the server struct to work with
	server = fshandler.Server{
		BaseDirectory: *baseDirectory,
		CertFile:      *certPath,
		KeyFile:       *keyPath,
		JWTCertFile:   *jwtCertPath,
	}

	// Validate our server config.
	err := server.Validate()
	if err != nil {
		panic(err)
	}

	pemBytes := bytes.NewBuffer(nil)

	certFile, err := os.Open(*jwtCertPath)
	if err != nil {
		log.Fatalf("Unale to open certificate file: %s, error: %s", *jwtCertPath, err)
	}

	_, err = io.Copy(pemBytes, certFile)
	if err != nil {
		log.Fatalf("Unable to extract certificate: %s", err)
	}

	// Get a closure to use with the jwt stuff.
	keyFunc, err := jwtclient.KeyFuncFromPEMBytes(pemBytes.Bytes())
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

	chain := alice.New(timeoutHandler, recoveryHandler, loggingHandler, jwthandler.Handler)

	// CoatLocker
	router.HandleFunc("/health", server.HealthEndpoint).Methods("GET")
	router.PathPrefix("/").Handler(chain.ThenFunc(server.GetEndpoint)).Methods("GET")
	router.PathPrefix("/").Handler(chain.ThenFunc(server.PutEndpoint)).Methods("PUT")
	router.PathPrefix("/").Handler(chain.ThenFunc(server.DeleteEndpoint)).Methods("DELETE")

	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(*listenAddress, *listenPort), *certPath, *keyPath, router))

}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 90*time.Second, "timed out")
}

func recoveryHandler(h http.Handler) http.Handler {
	return hndl.RecoveryHandler()(h)
}

func loggingHandler(h http.Handler) http.Handler {
	return hndl.LoggingHandler(os.Stdout, h)
}
