package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

func main() {

	c := new(dns.Client)

	http.HandleFunc("/dns-query", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		m := new(dns.Msg)
		err = m.Unpack(body)
		if err != nil {
			panic(err)
		}
		in, rtt, err := c.Exchange(m, "1.1.1.1:53")
		if err != nil {
			panic(err)
		}
		fmt.Println(rtt)
		b, err := in.Pack()
		if err != nil {
			panic(err)
		}
		w.Write(b)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
