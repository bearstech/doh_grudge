package doh

import (
	"fmt"
	"io/ioutil"
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
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	s.lock.RLock()
	cc, ok := s.cache[k]
	if ok {
		r := cc.Copy()
		r.Id = m.Id
		b, err := r.Pack()
		if err != nil {
			panic(err)
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
		panic(err)
	}
	fmt.Println(rtt)
	b, err := in.Pack()
	if err != nil {
		panic(err)
	}
	s.lock.Lock()
	s.cache[k] = in
	s.lock.Unlock()
	w.Write(b)
}
