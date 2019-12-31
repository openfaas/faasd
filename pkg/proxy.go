package pkg

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"time"
)

func NewProxy(timeout time.Duration) *Proxy {

	return &Proxy{
		Timeout: timeout,
	}
}

type Proxy struct {
	Timeout time.Duration
}

func (p *Proxy) Start(gatewayChan chan string) error {
	tcp := 8080

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	data := struct{ host string }{
		host: "",
	}

	data.host = <-gatewayChan

	log.Printf("Starting faasd proxy on %d\n", tcp)

	fmt.Printf("Gateway: %s\n", data.host)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcp),
		ReadTimeout:    p.Timeout,
		WriteTimeout:   p.Timeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			query := ""
			if len(r.URL.RawQuery) > 0 {
				query = "?" + r.URL.RawQuery
			}

			upstream := fmt.Sprintf("http://%s:8080%s%s", data.host, r.URL.Path, query)
			fmt.Printf("[faasd] proxy: %s\n", upstream)

			if r.Body != nil {
				defer r.Body.Close()
			}

			wrapper := ioutil.NopCloser(r.Body)
			upReq, upErr := http.NewRequest(r.Method, upstream, wrapper)

			copyHeaders(upReq.Header, &r.Header)

			if upErr != nil {
				log.Println(upErr)

				http.Error(w, upErr.Error(), http.StatusInternalServerError)
				return
			}

			upRes, upResErr := http.DefaultClient.Do(upReq)

			if upResErr != nil {
				log.Println(upResErr)

				http.Error(w, upResErr.Error(), http.StatusInternalServerError)
				return
			}

			copyHeaders(w.Header(), &upRes.Header)

			w.WriteHeader(upRes.StatusCode)
			io.Copy(w, upRes.Body)

		}),
	}

	return s.ListenAndServe()
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}
