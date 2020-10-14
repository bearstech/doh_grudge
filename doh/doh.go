package doh

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/miekg/dns"
)

type Server struct {
	cache    map[string]*dns.Msg
	lock     *sync.RWMutex
	resolver string
	client   *dns.Client
}

func New(resolver string) *Server {
	return &Server{
		cache:    make(map[string]*dns.Msg),
		lock:     &sync.RWMutex{},
		resolver: resolver,
		client:   new(dns.Client),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body []byte
	var err error
	if r.Method == "GET" {
		params, ok := r.URL.Query()["dns"]
		if !ok || len(params) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = base64.RawURLEncoding.Decode([]byte(params[0]), body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	m := new(dns.Msg)
	err = m.Unpack(body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mm := m.Copy()
	mm.Id = 0
	k := mm.String()
	fmt.Println(k)
	s.lock.RLock()
	cc, ok := s.cache[k]
	if ok {
		r := cc.Copy()
		r.Id = m.Id
		b, err := r.Pack()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
		s.lock.RUnlock()
		fmt.Println("Hit")
		return
	}
	s.lock.RUnlock()
	fmt.Println("Miss")
	in, rtt, err := s.client.Exchange(m, s.resolver)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(rtt)
	b, err := in.Pack()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.lock.Lock()
	s.cache[k] = in
	s.lock.Unlock()
	w.Header().Add("Content-Type", "application/dns-message")
	w.Write(b)
}
