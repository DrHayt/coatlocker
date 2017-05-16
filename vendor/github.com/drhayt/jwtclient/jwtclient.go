package jwtclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

// Client is the basic struct for this client.
type Client struct {
	username     string
	password     string
	url          string
	insecure     bool
	jwtCertBytes []byte
	keyFunc      jwt.Keyfunc
	token        *jwt.Token
}

// New takes a config struct, and returns a client struct.
func New(cfg *Config) (*Client, error) {
	if len(cfg.JWTCertBytes) == 0 {
		return nil, fmt.Errorf("Invalid certificate (length == 0?)")

	}

	if len(cfg.URL) == 0 {
		return nil, fmt.Errorf("Invalid URl (length == 0?)")
	}

	if len(cfg.Username) == 0 {
		return nil, fmt.Errorf("Invalid username (length == 0?)")
	}

	if len(cfg.Password) == 0 {
		return nil, fmt.Errorf("Invalid password (length == 0?)")
	}

	client := Client{
		insecure:     cfg.Insecure,
		username:     cfg.Username,
		password:     cfg.Password,
		url:          cfg.URL,
		jwtCertBytes: cfg.JWTCertBytes,
	}

	keyFunc, err := KeyFuncFromPEMBytes(client.jwtCertBytes)
	if err != nil {
		return nil, err
	}

	client.keyFunc = keyFunc

	return &client, nil

}

// Authenticate is a jwt wrapper that returns the JWT token to be used on subsequent calls.
func (c *Client) Authenticate() (err error) {

	// Setup the request
	request, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return
	}

	// Set our auth parameters.
	request.SetBasicAuth(c.username, c.password)

	// Get a config to override potential SSL stuff.
	tlsConfig := &tls.Config{}

	// If we are insecure, then so be it.
	if c.insecure {
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

	switch webResponse.StatusCode {
	case http.StatusOK:
		break
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized")
	case http.StatusInternalServerError:
		return fmt.Errorf("server Error")
	default:
		return fmt.Errorf("unexpected code received: %s", webResponse.Status)
	}

	// Get the body
	buffer := &bytes.Buffer{}
	io.Copy(buffer, webResponse.Body)

	token, err := jwt.Parse(buffer.String(), c.keyFunc)
	if err != nil {
		return err
	}

	c.token = token
	return nil

}

// StillValid checks to see if the token we have is still valid.
func (c *Client) StillValid() bool {
	if c.token == nil {
		return false
	}

	token, err := jwt.Parse(c.token.Raw, c.keyFunc)
	if err != nil {
		return false
	}

	if token.Valid {
		return true
	}

	return false

}

// RetrieveToken checks to see if the token we have is still valid, re-authenticates if necessary and returns the token.
func (c *Client) RetrieveToken() (string, error) {

	// Is our token still valid?
	if c.StillValid() {
		return c.token.Raw, nil
	}

	// Must re-authenticate
	err := c.Authenticate()
	if err != nil {
		return "", fmt.Errorf("unable to authenticate: %s", err)
	}

	return c.token.Raw, nil
}

// KeyFuncFromPEMBytes returns a closure which in turn returns the public part of the JWT certificate from a file.
func KeyFuncFromPEMBytes(pemBytes []byte) (jwt.Keyfunc, error) {

	// Decode the certificate
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("unable to decode pem encoded certificate")
	}

	pub, err := x509.ParseCertificate(block.Bytes)
	// pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: " + err.Error())
	}

	return func(*jwt.Token) (interface{}, error) { return pub.PublicKey, nil }, nil
}

// KeyFuncFromPEMFile returns a closure which in turn returns the public part of the JWT certificate from a file.
func KeyFuncFromPEMFile(fileLocation string) (jwt.Keyfunc, error) {

	pemFile, err := os.Open(fileLocation)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file \"%s\", error: %s", fileLocation, err)
	}

	defer pemFile.Close()

	pemBytes := bytes.NewBuffer(nil)

	_, err = io.Copy(pemBytes, pemFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file \"%s\", error: %s", fileLocation, err)
	}

	return KeyFuncFromPEMBytes(pemBytes.Bytes())

}
