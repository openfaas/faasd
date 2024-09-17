package pkg

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ConnectivityCheck checks if the controller can reach the
// public Internet via HTTPS.
// A license is required to use OpenFaaS CE for Commercial Use.
func ConnectivityCheck() error {
	req, err := http.NewRequest(http.MethodGet, "https://checkip.amazonaws.com", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("openfaas-ce/%s faas-netes", Version))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		var body []byte
		if res.Body != nil {
			body, _ = io.ReadAll(res.Body)
		}

		return fmt.Errorf("unexpected status code checking connectivity: %d, body: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
