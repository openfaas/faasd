package function

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		// read request payload
		reqBody, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		input = reqBody
		}
	}

	// log to stdout
	fmt.Printf("request body: %s", string(input))

	response := struct {
		Payload     string              `json:"payload"`
		Headers     map[string][]string `json:"headers"`
		Environment []string            `json:"environment"`
	}{
		Payload:     string(input),
		Headers:     r.Header,
		Environment: os.Environ(),
	}

	resBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // write result
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)
}