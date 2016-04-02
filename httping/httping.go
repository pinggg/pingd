package httping

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Timeout sets the ping timeout in milliseconds
var TimeOut = 5 * time.Second

// Ping sends a HEAD command to a given URL, returns whether the host answers 200 or not
func Ping(url string) (up bool, err error) {
	client := http.Client{
		Timeout: TimeOut,
	}

	resp, err := client.Head(url)
	if err != nil {
		return false, err
	}

	// Drain body just in case server misbehaves
	defer func() {
		n, _ := io.Copy(ioutil.Discard, resp.Body)
		if n > 0 {
			log.Printf("warning: received %d bytes on response body for url %s", n, url)
		}

		resp.Body.Close()
	}()

	if resp.StatusCode == 200 {
		return true, nil
	}

	return false, errors.New(resp.Status)
}
