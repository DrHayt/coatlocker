package jwtclient

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {

	insecure := flag.Bool("i", false, "Insecure, dont validate certificates")
	username := flag.String("user", "admin", "What username to use")
	password := flag.String("pass", "admin", "What password to use")
	url := flag.String("url", "https://authentication.sgtec.io", "Url of authentication service")

	flag.Parse()

	for i := 0; i <= 10; i++ {
		output, err := Authenticate(*insecure, *url, *username, *password)
		if err != nil {
			log.Println("Error authenticating: ", err)
			continue
		}
		fmt.Printf("Response was %s\n", output)
	}

}

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
