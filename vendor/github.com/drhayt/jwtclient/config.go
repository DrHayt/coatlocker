package jwtclient

// Config is the package configuration struct that configures the package.
type Config struct {
	Username     string
	Password     string
	URL          string
	Insecure     bool
	JWTCertBytes []byte
}
