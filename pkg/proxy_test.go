package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func Test_Proxy_ToPrivateServer(t *testing.T) {

	wantBodyText := "OK"
	wantBody := []byte(wantBodyText)
	upstreamSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		w.WriteHeader(http.StatusOK)
		w.Write(wantBody)

	}))

	defer upstreamSvr.Close()
	port := 8080
	u, _ := url.Parse(upstreamSvr.URL)
	log.Println("Host", u.Host)

	upstreamAddr := u.Host
	proxy := NewProxy(upstreamAddr, 8080, "127.0.0.1", time.Second*1, &mockResolver{})

	gwChan := make(chan string, 1)
	doneCh := make(chan bool)

	go proxy.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		gwChan <- u.Host
		wg.Done()
	}()
	wg.Wait()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d", port), nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < 11; i++ {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Logf("Try %d, gave error: %s", i, err)

			time.Sleep(time.Millisecond * 100)
		} else {

			resBody, _ := ioutil.ReadAll(res.Body)
			if string(resBody) != string(wantBody) {
				t.Errorf("want %s, but got %s in body", string(wantBody), string(resBody))
			}
			break
		}
	}

	go func() {
		doneCh <- true
	}()
}

type mockResolver struct {
}

func (m *mockResolver) Start() {

}

func (m *mockResolver) Get(upstream string, got chan<- string, timeout time.Duration) {
	got <- upstream
}
