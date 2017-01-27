package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	err         error
	envTarget   string // target URL read from an environment variable
	envPort     string // port over which to run this reverse proxy service
	envKey      string // header key to be authenticated
	envKeyVal   string // header key value expected to be found to successfully authenticate
	rprxHandler *httputil.ReverseProxy
	target      *url.URL
)

// init function configures operating variables
func init() {
	envTarget = os.Getenv("RP_TARGET_URL")
	envPort = os.Getenv("RP_PORT")
	envKey = os.Getenv("RP_HEADER_KEY")
	envKeyVal = os.Getenv("RP_HEADER_KEY_VAL")

	// Verify that environment variables were found
	if len(envPort) == 0 || len(envTarget) == 0 || len(envKey) == 0 || len(envKeyVal) == 0 {
		log.Printf("ERROR: at least one required environment variable is missing: %v. See: %v\n", envTarget, err)
		os.Exit(1)
	}
}

// authMidWare is a Middleware function that authenticates the inbound request
//  requests with the correct header information will continue to the target server
//  requests with incorrect header information will return as unauthorized
func authMidWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Middleware logic to authenticate the request
		if r.Header.Get("X-ContentKey") == "OK" {
			// Correct API Key found, proceed with the next handler
			next.ServeHTTP(w, r)
		} else {
			// Incorrect API Key, respond as unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized request"))
		}
	})
}

func main() {
	// Create a URL object to represent the reverse proxy target address
	if target, err = url.Parse(envTarget); err != nil {
		// Error occurred parsing the target url
		log.Printf("ERROR: error occurred parsing the target url: %v. See: %v\n", envTarget, err)
		os.Exit(1)
	}

	// Create a new reverse proxy handler
	rprxHandler = httputil.NewSingleHostReverseProxy(target)

	// Register a handler with the base route that consists of a...
	//   1. wrapping function that authenticates the request
	//   2. wrapped function that handles authenticated requests and proxies the request to the target
	http.Handle("/", authMidWare(rprxHandler))

	// Start listening on the specified port
	log.Printf("INFO: listening on port: %v and targeting: %v\n", envPort, envTarget)
	http.ListenAndServe(envPort, nil)
}
