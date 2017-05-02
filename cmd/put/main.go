package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/drhayt/coatlocker/pkg/jwtclient"
)

func main() {

	insecure := flag.Bool("i", false, "Insecure, dont validate certificates")
	username := flag.String("user", "admin", "What username to use")
	password := flag.String("pass", "admin", "What password to use")
	url := flag.String("url", "https://authentication.sgtec.io", "Url of authentication service")

	flag.Parse()

	output, err := jwtclient.Authenticate(*insecure, *url, *username, *password)
	if err != nil {
		log.Fatalln("Error authenticating: ", err)
	}
	fmt.Printf(output)
}
