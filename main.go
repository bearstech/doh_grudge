package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/miekg/dns"
)

func main() {

	c := new(dns.Client)
	cache := make(map[string]*dns.Msg)
	l := sync.RWMutex{}

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
		mm := m.Copy()
		mm.Id = 0
		k := mm.String()
		fmt.Println(k)
		l.RLock()
		cc, ok := cache[k]
		if ok {
			r := cc.Copy()
			r.Id = m.Id
			b, err := r.Pack()
			if err != nil {
				panic(err)
			}
			w.Write(b)
			l.RUnlock()
			fmt.Println("Hit")
			return
		}
		l.RUnlock()
		fmt.Println("Miss")
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
		cache[k] = in
		l.Unlock()
		w.Write(b)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
