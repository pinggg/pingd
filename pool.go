package pingd

import (
	"log"
	"time"
)

// Host is a wrap around a host (name or IP) and the host status
// represented by Down.  The status is used as initial state when
// monitoring starts and a event information when a host goes
// up or down.
type Host struct {
	Host string
	Down bool
}

// Receiver is a functions which takes 2 channels of Host
// in the first ones inserts Host(s) that should be monitored
// in the second one Host(s) that should stop being monitored
type Receiver func(chan<- Host, chan<- Host)

// Notifier is a function which takes 1 channel of Host(s)
// where it all hosts that went throw an UP or DOWN status change.
type Notifier func(<-chan Host)

// Loader is a function which takes 1 channel of Host(s)
// where it should insert Host(s) that should be monitored
// this function will run at boot time to load an initial
// list of Host(s)
type Loader func(chan<- Host)

// Pool is the structure that wraps the list of Host(s) that are
// being monitored, with the monitoring parameters and the functions
// interfacing with the rest of the system.
type Pool struct {
	Interval  time.Duration
	FailLimit int
	Receive   Receiver
	Notify    Notifier
	Load      Loader

	list map[string]*Monitor
}

// Start create the necessary internal channels and
// calls all necessary functions to start the engine.
func (p *Pool) Start() {
	p.list = make(map[string]*Monitor)
	startHostCh := make(chan Host, 10)
	stopHostCh := make(chan Host, 10)
	notifyCh := make(chan Host, 10)

	if p.Load != nil {
		go p.Load(startHostCh)
	}

	if p.Notify != nil {
		go p.Notify(notifyCh)
	}

	if p.Receive != nil {
		go p.Receive(startHostCh, stopHostCh)
	}

	go p.run(startHostCh, stopHostCh, notifyCh)
}

// run glues together the channels for communication with the host monitors
// and the rest of the system.
func (p *Pool) run(startHostCh, stopHostCh <-chan Host, notifyCh chan<- Host) {
	for {
		select {

		// START
		case h := <-startHostCh:

			if _, exists := p.list[h.Host]; exists {
				log.Println("RESTART pinging " + h.Host)
				go func(h *Monitor) {
					h.Stop()
					h.Start(p.Interval, p.FailLimit)
				}(p.list[h.Host])
			} else {
				log.Println("NEW host " + h.Host)
				p.list[h.Host] = NewMonitor(h, notifyCh)
				go func(h *Monitor) {
					h.Start(p.Interval, p.FailLimit)
				}(p.list[h.Host])
			}

			// STOP
		case h := <-stopHostCh:

			if _, exists := p.list[h.Host]; exists {
				log.Println("STOP pinging " + h.Host)
				p.list[h.Host].Stop()

			} else {
				log.Println("ERROR host not found " + h.Host)
			}
		}
	}
}
