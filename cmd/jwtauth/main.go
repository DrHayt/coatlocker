package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/drhayt/jwtclient"
)

func main() {

	insecure := flag.Bool("i", false, "Insecure, dont validate certificates")
	username := flag.String("user", "admin", "What username to use")
	password := flag.String("pass", "admin", "What password to use")
	authurl := flag.String("authurl", "https://authentication.sgtec.io", "Url of authentication service")
	jwtCert := flag.String("jwtCertPath", "jwt.crt", "Certificate file to use to validate the jwt.")

	flag.Parse()

	jwtBytes, err := ioutil.ReadFile(*jwtCert)
	if err != nil {
		log.Fatalf("Unable to read JWT cert file: %s, Error: %s", *jwtCert, err)
	}

	jwtConfig := jwtclient.Config{
		Username:     *username,
		Password:     *password,
		URL:          *authurl,
		Insecure:     *insecure,
		JWTCertBytes: jwtBytes,
	}

	jwtClient, err := jwtclient.New(&jwtConfig)
	if err != nil {
		log.Fatalf("Unable to create jwt client: %s", err)
	}

	token, err := jwtClient.RetrieveToken()
	if err != nil {
		log.Fatalln("Error authenticating: ", err)
	}

	fmt.Println(token)
}
