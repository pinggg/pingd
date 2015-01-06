package std

import (
	"log"

	"github.com/pinggg/pingd"
)

// NewLoaderFunc gets the sequence of hostnames/IPs and returns a Loader
// function that when called will insert them as Host structs into the
// start channel at boot time.
func NewLoaderFunc(hosts []string) pingd.Loader {
	return func(load chan<- pingd.Host) {
		for _, host := range hosts {
			load <- pingd.Host{Host: host, Down: false}
		}
	}
}

// NewNotifierFunc returns a function with just logs up and down events
func NewNotifierFunc() pingd.Notifier {
	return func(notifyCh <-chan pingd.Host) {
		for {
			select {
			case h := <-notifyCh:
				switch h.Down {
				case true:
					log.Println("DOWN " + h.Host)
				case false:
					log.Println("UP " + h.Host)
				}
			}
		}
	}
}
