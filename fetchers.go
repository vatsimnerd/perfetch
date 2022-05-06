package perfetch

import (
	"io/ioutil"
	"net/http"
	"time"
)

func HTTPGetFetcher(url string, timeout time.Duration) Fetcher[[]byte] {
	return func() ([]byte, error) {
		cli := &http.Client{Timeout: timeout}
		resp, err := cli.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
}
