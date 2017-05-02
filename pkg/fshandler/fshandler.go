package fshandler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	respond "gopkg.in/matryer/respond.v1"
)

// Server is the struct that represents the server.
type Server struct {
	BaseDirectory string
}

// HealthEndpoint is an endpoint to allow for health monitoring.
func (s *Server) HealthEndpoint(w http.ResponseWriter, r *http.Request) {
	respond.WithStatus(w, r, http.StatusOK)

}

// DeleteEndpoint handles deleting a file if it exists.
func (s *Server) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.genKey(r.RequestURI)
	filepath := s.genPath(key)

	// Dont try to get a file that does not exists.
	err := checkFile(filepath)
	if err != nil {
		respond.WithStatus(w, r, http.StatusNotFound)
		return
	}

	// Ok, its there, actually remove it.
	err = os.Remove(filepath)
	if err != nil {
		respond.WithStatus(w, r, http.StatusInternalServerError)
		return
	}
	respond.WithStatus(w, r, http.StatusOK)
}

// GetEndpoint is the endpoint that does stuff.
func (s *Server) GetEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.genKey(r.RequestURI)
	filepath := s.genPath(key)

	// Dont try to get a file that does not exists.
	err := checkFile(filepath)
	if err != nil {
		respond.WithStatus(w, r, http.StatusNotFound)
		return
	}

	// It must exist, so open it up.
	file, err := os.Open(filepath)
	if err != nil {
		respond.WithStatus(w, r, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Stuff must be good.
	w.WriteHeader(http.StatusOK)
	// Lets stream some bytes.
	// _, err = io.Copy(w, file)
	// we dont care if this fails.
	// respond.With(w, r, http.StatusInternalServerError, file.
	io.Copy(w, file)
	return
}

// PutEndpoint is the endpoint that does stuff.
func (s *Server) PutEndpoint(w http.ResponseWriter, r *http.Request) {

	key := s.genKey(r.RequestURI)
	filepath := s.genPath(key)

	// Dont try to put a file that already exists.
	err := checkFile(filepath)
	if err == nil {
		respond.WithStatus(w, r, http.StatusUnprocessableEntity)
		return
	}

	file, err := os.Create(filepath)
	if err != nil {
		respond.WithStatus(w, r, http.StatusInternalServerError)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, r.Body)
	if err != nil {
		respond.WithStatus(w, r, http.StatusInternalServerError)
		return
	}
	respond.WithStatus(w, r, http.StatusCreated)

}

// Validate validates that the server is proper.
func (s *Server) Validate() error {
	err := checkDir(s.BaseDirectory)
	if err != nil {
		return err
	}
	return nil
}

// Genkey generates a key to be used inside of other function.s
func (s Server) genKey(URI string) string {
	hasher := sha256.New()
	hasher.Write([]byte(URI))
	key := hex.EncodeToString(hasher.Sum(nil))
	return (key)
}

func (s Server) genPath(Key string) string {
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
