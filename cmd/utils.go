package cmd

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
)

// getContentType resolves the correct Content-Type for a proxied function.
func getContentType(request http.Header, proxyResponse http.Header) (headerContentType string) {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultContentType
	}

	return headerContentType
}

func hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func captureRequestData(req *http.Request) ([]byte, error) {
	var b = &bytes.Buffer{} // holds serialized representation
	//var tmp *http.Request
	var err error
	if err = req.Write(b); err != nil { // serialize request to HTTP/1.1 wire format
		return nil, err
	}
	//var reqSerialize []byte

	return b.Bytes(), nil
	//r := bufio.NewReader(b)
	//if tmp, err = http.ReadRequest(r); err != nil { // deserialize request
	//	return nil,err
	//}
	//*req = *tmp // replace original request structure
	//return nil
}

func unserializeReq(sReq []byte, req *http.Request) (*http.Response, error) {
	b := bytes.NewBuffer(sReq)
	r := bufio.NewReader(b)
	res, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		// vClone := make([]string, len(v))
		// var vClone []string
		vClone := v
		// copy(vClone[:], v)

		destination[k] = vClone
	}
}
