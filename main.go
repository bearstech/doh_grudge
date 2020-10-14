package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/miekg/dns"
)

func main() {

	c := new(dns.Client)
	cache := make(map[string][]byte)
	fmt.Println(cache)
	l := sync.RWMutex{}

	http.HandleFunc("/dns-query", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		encoded := base64.StdEncoding.EncodeToString(body)
		l.RLock()
		cc, ok := cache[encoded]
		if ok {
			w.Write(cc)
			l.RUnlock()
			return
		}
		l.RUnlock()
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
		l.Lock()
		cache[encoded] = b
		l.Unlock()
		w.Write(b)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
