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
	// router.HandleFunc("/API/user/{id}", repo.WithData(handlers.UpdateUserEndpoint)).Methods("PUT")
	// router.HandleFunc("/API/user/{id}", repo.WithData(handlers.DeleteUserEndpoint)).Methods("DELETE")
	go http.ListenAndServe(fmt.Sprintf("%s:%d", *listenAddress, *listenPort), h)
	// You could do this on like anything.
	log.Fatal(http.ListenAndServe(":8887", nil)) // For net/trace
}

func (s *Server) HealthEndpoint(w http.ResponseWriter, r *http.Request) {
	respond.With(w, r, http.StatusOK, s.BaseDirectory)

}

func (s *Server) DeleteFileEndpoint(w http.ResponseWriter, r *http.Request) {

	hasher := sha256.New()
	hasher.Write([]byte(r.RequestURI))
	key := hex.EncodeToString(hasher.Sum(nil))

	err := os.Remove(path.Join(s.BaseDirectory, key))
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}
	respond.With(w, r, http.StatusOK, key)
}

func (s *Server) GetFileEndpoint(w http.ResponseWriter, r *http.Request) {

	hasher := sha256.New()
	hasher.Write([]byte(r.RequestURI))
	key := hex.EncodeToString(hasher.Sum(nil))

	file, err := os.Open(path.Join(s.BaseDirectory, key))
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, file)
	if err != nil {
		respond.With(w, r, http.StatusInternalServerError, err.Error())
	}

}

func (s *Server) PutFileEndpoint(w http.ResponseWriter, r *http.Request) {

	hasher := sha256.New()
	hasher.Write([]byte(r.RequestURI))
	key := hex.EncodeToString(hasher.Sum(nil))

	file, err := os.Create(path.Join(s.BaseDirectory, key))
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

func checkFile(path string) error {
	DirStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if DirStat.IsDir() {
		return errors.New(fmt.Sprintf("%s: %s", path, "is not a file"))
	}
	return nil
}
