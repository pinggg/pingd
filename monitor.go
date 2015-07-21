package pingd

import (
	"log"
	"sync"
	"time"

	"github.com/pinggg/pingd/ping"
)

var pingFunc = ping.Ping // ping function

// Monitor is the main structure that represent a monitored host
// Whenever a host goes up or down it notifies it on the corresponding channel
type Monitor struct {
	host      Host
	failures  int
	failLimit int
	interval  time.Duration
	stop      bool
	notifyCh  chan<- Host
	running   *sync.Mutex // monitor must run only once
	lock      *sync.Mutex // protects internal values
}

// NewMonitor takes a host, an initial state, and the notification channels and returns a monitorable host structure
func NewMonitor(host Host, notifyCh chan<- Host) *Monitor {
	h := Monitor{
		host:     host,
		notifyCh: notifyCh,
		running:  &sync.Mutex{},
		lock:     &sync.Mutex{},
	}

	return &h
}

// Start begins the periodic pinging of the host
func (m *Monitor) Start(interval time.Duration, failLimit int) {
	m.running.Lock()
	defer m.running.Unlock()
	defer m.lock.Unlock()

	m.lock.Lock()
	m.interval = interval
	m.failLimit = failLimit
	m.stop = false
	m.lock.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for _ = range ticker.C {
		//		log.Println("tick")
		m.lock.Lock()
		if m.stop {
			return
		}
		m.lock.Unlock()

		if up, err := pingFunc(m.host.Host); up {
			//			log.Println(m.host.Host + " pong")
			m.markUp()
		} else {
			//			log.Println(m.host.Host + " failed")
			m.markDown(err)
		}
	}
}

// Stop stops pinging the host
func (m *Monitor) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.stop = true
}

// markUp resets the failure count and the host status, then sends a channel notification that the host is up.
func (m *Monitor) markUp() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.host.Down {
		m.failures = 0
		return
	}

	m.failures--
	if m.failures > 0 {
		return
	}

	m.host.Down = false
	m.notifyCh <- m.host
}

// markDown does nothing if the host is already down. If it's up, in increases the failure count
// changes the status to down and then sends a channel notification that the host is down.
func (m *Monitor) markDown(err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.host.Down {
		m.failures = m.failLimit
		return
	}

	m.failures++
	if m.failures < m.failLimit {
		return
	}

	if err != nil {
		log.Println(err.Error())
	}

	m.host.Down = true
	m.notifyCh <- m.host
}
