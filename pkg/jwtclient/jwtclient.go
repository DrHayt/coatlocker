package jwtclient

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
)

// Authenticate is a jwt wrapper that returns the JWT token to be used on subsequent calls.
func Authenticate(insecure bool, url, username, password string) (token string, err error) {

	// Setup the request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Set our auth parameters.
	request.SetBasicAuth(username, password)

	// Get a config to override potential SSL stuff.
	tlsConfig := &tls.Config{}

	// If we are insecure, then so be it.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Create the transport
	tr := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a client using that transport.
	client := &http.Client{Transport: tr}

	// Make the request
	webResponse, err := client.Do(request)
	if err != nil {
		return
	}

	defer webResponse.Body.Close()
	buffer := &bytes.Buffer{}
	io.Copy(buffer, webResponse.Body)
	token = buffer.String()
	return
}

// RetrieveCertificate is a wrapper to retrieve a certificate from a remote server.
func RetrieveCertificate(insecure bool, url string) (certificate string, err error) {

	// Setup the request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Get a config to override potential SSL stuff.
	tlsConfig := &tls.Config{}

	// If we are insecure, then so be it.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Create the transport
	tr := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a client using that transport.
	client := &http.Client{Transport: tr}

	// Make the request
	webResponse, err := client.Do(request)
	if err != nil {
		return
	}

	defer webResponse.Body.Close()
	buffer := &bytes.Buffer{}
	io.Copy(buffer, webResponse.Body)
	certificate = buffer.String()
	return
}
