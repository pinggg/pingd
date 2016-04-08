package pingd

import (
	"sync"
	"time"
)

// PingFunc is function signature for ping checks
type PingFunc func(host string) (up bool, err error)

// Monitor is the main structure that represent a monitored host
// Whenever a host goes up or down it notifies it on the corresponding channel
type Monitor struct {
	running   *sync.Mutex // monitor must run only once
	lock      *sync.Mutex // protects internal values
	ping      PingFunc
	host      string
	down      bool
	failures  int
	failLimit int
	interval  time.Duration
	stop      bool
	notifyCh  chan<- HostStatus
}

// NewMonitor takes a host, an initial state, and the notification channels and returns a monitorable host structure
func NewMonitor(status HostStatus, ping PingFunc, notifyCh chan<- HostStatus) *Monitor {
	h := Monitor{
		ping:     ping,
		host:     status.Host,
		down:     status.Down,
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

		if up, err := m.ping(m.host); up {
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
	if !m.down {
		m.failures = 0
		return
	}

	m.failures--
	if m.failures > 0 {
		return
	}

	m.down = false
	m.notifyCh <- HostStatus{Host: m.host, Down: m.down}
}

// markDown does nothing if the host is already down. If it's up, in increases the failure count
// changes the status to down and then sends a channel notification that the host is down.
func (m *Monitor) markDown(err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.down {
		m.failures = m.failLimit
		return
	}

	m.failures++
	if m.failures < m.failLimit {
		return
	}

	m.down = true
	m.notifyCh <- HostStatus{Host: m.host, Down: m.down, Reason: err}
}
