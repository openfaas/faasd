package pkg

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"time"
)

// NewProxy creates a HTTP proxy to expose a host
func NewProxy(upstream string, listenPort uint32, hostIP string, timeout time.Duration, resolver Resolver) *Proxy {

	return &Proxy{
		Upstream: upstream,
		Port:     listenPort,
		HostIP:   hostIP,
		Timeout:  timeout,
		Resolver: resolver,
	}
}

// Proxy for exposing a private container
type Proxy struct {
	Timeout time.Duration

	// Port on which to listen to traffic
	Port uint32

	// Upstream is where to send traffic when received
	Upstream string

	// The IP to use to bind locally
	HostIP string

	Resolver Resolver
}

// Start listening and forwarding HTTP to the host
func (p *Proxy) Start() error {

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	upstreamHost, upstreamPort, err := getUpstream(p.Upstream, p.Port)
	if err != nil {
		return err
	}

	log.Printf("Looking up IP for: %q", upstreamHost)
	got := make(chan string, 1)

	go p.Resolver.Get(upstreamHost, got, time.Second*5)

	ipAddress := <-got
	close(got)

	upstreamAddr := fmt.Sprintf("%s:%d", ipAddress, upstreamPort)

	localBind := fmt.Sprintf("%s:%d", p.HostIP, p.Port)
	log.Printf("Proxy from: %s, to: %s (%s)\n", localBind, p.Upstream, ipAddress)

	l, err := net.Listen("tcp", localBind)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return err
	}

	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			acceptErr := fmt.Errorf("unable to accept on %d, error: %s",
				p.Port,
				err.Error())
			log.Printf("%s", acceptErr.Error())
			return acceptErr
		}

		upstream, err := net.Dial("tcp", upstreamAddr)
		if err != nil {
			log.Printf("unable to dial to %s, error: %s", upstreamAddr, err.Error())
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

func getUpstream(val string, defaultPort uint32) (string, uint32, error) {
	upstreamHostname := val
	upstreamPort := defaultPort

	if in := strings.Index(val, ":"); in > -1 {
		upstreamHostname = val[:in]
		port, err := strconv.ParseInt(val[in+1:], 10, 32)
		if err != nil {
			return "", defaultPort, err
		}
		upstreamPort = uint32(port)
	}

	return upstreamHostname, upstreamPort, nil
}
