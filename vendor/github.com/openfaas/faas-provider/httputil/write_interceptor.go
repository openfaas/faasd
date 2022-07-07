package httputil

import (
	"bufio"
	"net"
	"net/http"
)

func NewHttpWriteInterceptor(w http.ResponseWriter) *HttpWriteInterceptor {
	return &HttpWriteInterceptor{w, 0}
}

type HttpWriteInterceptor struct {
	http.ResponseWriter
	StatusCode int
}

func (c *HttpWriteInterceptor) Status() int {
	if c.StatusCode == 0 {
		return http.StatusOK
	}
	return c.StatusCode
}

func (c *HttpWriteInterceptor) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *HttpWriteInterceptor) Write(data []byte) (int, error) {
	return c.ResponseWriter.Write(data)
}

func (c *HttpWriteInterceptor) WriteHeader(code int) {
	c.StatusCode = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *HttpWriteInterceptor) Flush() {
	fl := c.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (c *HttpWriteInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := c.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (c *HttpWriteInterceptor) CloseNotify() <-chan bool {
	notifier, ok := c.ResponseWriter.(http.CloseNotifier)
	if ok == false {
		return nil
	}
	return notifier.CloseNotify()
}
