package main

import "net/http"

type ICoatHandler interface {
	Validate() error
	HealthEndpoint(w http.ResponseWriter, r *http.Request)
	GetEndpoint(w http.ResponseWriter, r *http.Request)
	PutEndpoint(w http.ResponseWriter, r *http.Request)
	DeleteEndpoint(w http.ResponseWriter, r *http.Request)
}
