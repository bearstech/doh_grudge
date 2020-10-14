package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bearstech/doh_grudge/doh"
)

func main() {

	resolver := os.Getenv("DNS")
	if resolver == "" {
		resolver = "1.1.1.1:53"
	}

	s := doh.New(resolver)

	http.Handle("/dns-query", s)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
