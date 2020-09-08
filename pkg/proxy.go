package pkg

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"time"
)

// NewProxy creates a HTTP proxy to expose the gateway container
// from OpenFaaS to the host
func NewProxy(port int, timeout time.Duration) *Proxy {

	return &Proxy{
		Port:    port,
		Timeout: timeout,
	}
}

// Proxy for exposing a private container
type Proxy struct {
	Timeout time.Duration
	Port    int
}

type proxyState struct {
	Host string
}

// Start listening and forwarding HTTP to the host
func (p *Proxy) Start(gatewayChan chan string, done chan bool) error {
	tcp := p.Port

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	ps := proxyState{
		Host: "",
	}

	ps.Host = <-gatewayChan

	log.Printf("Starting faasd proxy on %d\n", tcp)

	fmt.Printf("Gateway: %s\n", ps.Host)

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", tcp))
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return err
	}

	defer l.Close()

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			acceptErr := fmt.Errorf("Unable to accept on %d, error: %s", tcp, err.Error())
			log.Printf("%s", acceptErr.Error())
			return acceptErr
		}

		upstream, err := net.Dial("tcp", fmt.Sprintf("%s", ps.Host))

		if err != nil {
			log.Printf("unable to dial to %s, error: %s", ps.Host, err.Error())
			return err
		}

		go pipe(conn, upstream)
		go pipe(upstream, conn)

	}
}

func pipe(from net.Conn, to net.Conn) {
	defer from.Close()
	io.Copy(from, to)
}
