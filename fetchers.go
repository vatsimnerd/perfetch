package perfetch

import (
	"io/ioutil"
	"net/http"
	"time"

)

func HTTPGetFetcher(url string, timeout time.Duration) Fetcher[[]byte] {
	return func() ([]byte, error) {
		flog := log.WithField("url", url)
		flog.Debug("running HTTPGetFetcher")
		cli := &http.Client{Timeout: timeout}
		resp, err := cli.Get(url)
		if err != nil {
			flog.WithError(err).Error("error getting HTTP response")
			return nil, err
		}

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			flog.WithError(err).Error("error reading HTTP response body")
		}
		return data, err
	}
}
