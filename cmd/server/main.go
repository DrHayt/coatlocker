package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	respond "gopkg.in/matryer/respond.v1"

	hndl "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	BaseDirectory string
}

func main() {

	var (
		baseDirectory = flag.String("basedir", "/tmp", "The directory to use as the base of file uploads/downloads")
		listenPort    = flag.Int("port", 8443, "The port to listen on")
		listenAddress = flag.String("address", "127.0.0.1", "The address to listen on")
		router        = mux.NewRouter()
	)
	flag.Parse()

	server := Server{BaseDirectory: *baseDirectory}

	err := server.Validate()
	if err != nil {
		panic("Invalid base directory")
	}

	h := hndl.LoggingHandler(os.Stdout, router)

	//Users
	router.HandleFunc("/health", server.HealthEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.GetFileEndpoint).Methods("GET")
	router.PathPrefix("/").HandlerFunc(server.PutFileEndpoint).Methods("PUT")
	router.PathPrefix("/").HandlerFunc(server.DeleteFileEndpoint).Methods("DELETE")

	// This sets up the application to serve.
	go http.ListenAndServe(fmt.Sprintf("%s:%d", *listenAddress, *listenPort), h)

	// This sets up the tracing for requests.
	log.Fatal(http.ListenAndServe(":8887", nil)) // For net/trace

	// We should setup an additional GRPC endpoint.

}

// HealthEndpoint is an endpoint to allow for health monitoring.
func (s *Server) HealthEndpoint(w http.ResponseWriter, r *http.Request) {
	respond.With(w, r, http.StatusOK, s.BaseDirectory)

}

// DeleteFileEndpoint handles deleting a file if it exists.
func (s *Server) DeleteFileEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.GenKey(r.RequestURI)
	filepath := s.GenPath(key)

	err := os.Remove(filepath)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}
	respond.With(w, r, http.StatusOK, key)
}

func (s *Server) GetFileEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.GenKey(r.RequestURI)
	filepath := s.GenPath(key)

	// Dont try to get a file that does not exists.
	err := checkFile(filepath)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}

	// It must exist, so open it up.
	file, err := os.Open(filepath)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}

	// Stuff must be good.
	w.WriteHeader(http.StatusOK)
	// Lets stream some bytes.
	// _, err = io.Copy(w, file)
	// we dont care if this fails.
	io.Copy(w, file)
}

func (s *Server) PutFileEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.GenKey(r.RequestURI)
	filepath := s.GenPath(key)

	file, err := os.Create(filepath)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}
	written, err := io.Copy(file, r.Body)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}
	respond.With(w, r, http.StatusOK, fmt.Sprintf("Wrote %d bytes to %s", written, key))

}

func (s *Server) Validate() error {
	err := checkDir(s.BaseDirectory)
	if err != nil {
		return err
	}
	return nil
}

func (s Server) GenKey(URI string) string {
	hasher := sha256.New()
	hasher.Write([]byte(URI))
	key := hex.EncodeToString(hasher.Sum(nil))
	return (key)
}

func (s Server) GenPath(Key string) string {
	return path.Join(s.BaseDirectory, Key)
}

// Err unless a directory exists.
func verifyDirectoryExists(path string) error {
	DirStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !DirStat.IsDir() {
		return errors.New(fmt.Sprintf("%s: %s", path, "is not a directory"))
	}
	return nil
}

// Err unless a directory exists.
func checkDir(path string) error {
	DirStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !DirStat.IsDir() {
		return errors.New(fmt.Sprintf("%s: %s", path, "is not a directory"))
	}
	return nil
}

// Err unless a file exists.
func verifyFileExists(path string) error {
	FileStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if FileStat.IsDir() {
		return errors.New(fmt.Sprintf("%s: %s", path, "is not a file"))
	}
	return nil
}

// Err unless a file exists.
func checkFile(path string) error {
	FileStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if FileStat.IsDir() {
		return errors.New(fmt.Sprintf("%s: %s", path, "is not a file"))
	}
	return nil
}
