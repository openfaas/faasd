package pkg

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

	time.Sleep(3 * time.Second)
	log.Printf("Starting faasd proxy on %d\n", tcp)
	data := struct{ host string }{
		host: "",
	}

	fileData, fileErr := ioutil.ReadFile(p.Hosts)
	if fileErr != nil {
		return fileErr
	}

	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		if strings.Index(line, "gateway") > -1 {
			data.host = line[:strings.Index(line, "\t")]
		}
	}
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
			fmt.Printf("Forward to %s %s\n", upstream, data)

			if r.Body != nil {
				defer r.Body.Close()
			}

			wrapper := ioutil.NopCloser(r.Body)
			upReq, upErr := http.NewRequest(r.Method, upstream, wrapper)

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

			for k, v := range upRes.Header {
				w.Header().Set(k, v[0])
			}

			w.WriteHeader(upRes.StatusCode)
			io.Copy(w, upRes.Body)

		}),
	}

	return s.ListenAndServe()
}
