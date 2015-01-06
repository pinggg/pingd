package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pinggg/pingd"
)

type pingHTTP struct {
	startCh chan<- pingd.Host
	stopCh  chan<- pingd.Host
}

// ServeHTTP handles the incoming start/stop commands via HTTP
func (p pingHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Path[1:]
	if host == "" {
		fmt.Fprint(w, "missing host on request\n")
	}

	switch r.Method {
	case "DELETE":
		fmt.Fprintf(w, "stop ping %s\n", host)
		p.stopCh <- pingd.Host{Host: host, Down: false}
	default:
		fmt.Fprintf(w, "start ping %s\n", host)
		p.startCh <- pingd.Host{Host: host, Down: false}
	}
}

// NewReceiverFunc returns the functions with sets up the system channels
// and starts the webserver
func NewReceiverFunc(listen string) pingd.Receiver {
	return func(startCh, stopCh chan<- pingd.Host) {
		var p = &pingHTTP{startCh, stopCh}
		log.Printf("Web server starting on %s", listen)
		err := http.ListenAndServe(listen, p)
		if err != nil {
			log.Fatal(err)
		}

	}
}
