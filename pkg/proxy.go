package pkg

import (
	"fmt"
	"net/http"

	"time"
)

func NewProxy(hosts string, timeout time.Duration) *Proxy {

	return &Proxy{
		Hosts:   hosts,
		Timeout: timeout,
	}
}

type Proxy struct {
	Hosts   string
	Timeout time.Duration
}

func (p *Proxy) Start() error {
	tcp := 8080
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcp),
		ReadTimeout:    p.Timeout,
		WriteTimeout:   p.Timeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	return s.ListenAndServe()
}
