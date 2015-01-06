package pingd

import (
	"log"
	"sync"
	"testing"
	"time"
)

// TestUpDown tests correct event stream is coming from the monitoring pool
func TestUpDownStartStop(t *testing.T) {
	var sl SkipLog
	log.SetOutput(sl)

	seq := make(map[string][]bool)
	seq["h1"] = []bool{true, false, false, false, true, true, false}
	seq["h2"] = []bool{false, false, true, true, false, true, true}
	seq["h3"] = []bool{true, true, true, false, true, true, false}
	seq["h4"] = []bool{false, false, true, true, false, false, false}
	load := []string{"h1", "h2"}

	resultSeq := []Host{
		Host{"h2", true},  // h2 goes down first
		Host{"h1", true},  // h1 follows
		Host{"h2", false}, // h2 goes up
		Host{"h1", false}, // h1 goes up
		Host{"h1", true},  // h2 goes down
		Host{"h2", true},  // h1 goes down
	}

	startHostChFW, stopHostChFW, notifyChFW := createTestPool(seq, load)

	// Test expected events for h1 and h2
	for _, expected := range resultSeq {
		event := <-notifyChFW
		if event.Host != expected.Host || event.Down != expected.Down {
			t.Errorf("Got event: %s %t, expected: %s %t \n", event.Host, event.Down, expected.Host, expected.Down)
		}
	}

	startHostChFW <- Host{"h3", false} // start h3 as UP
	startHostChFW <- Host{"h4", true}  // start h4 as DOWN

	// Expect h4 to come UP (down=false) first
	event := <-notifyChFW
	if event.Host != "h4" || event.Down != false {
		t.Errorf("Got event: %s %t, expected: %s %t \n", event.Host, event.Down, "h4", false)
	}

	// Stop monitoring h4
	stopHostChFW <- event
	// Expect h3 will eventually go DOWN (down=true)
	event = <-notifyChFW
	if event.Host != "h3" || event.Down != true {
		t.Errorf("Got event: %s %t, expected: %s %t \n", event.Host, event.Down, "h3", true)
	}

	// There should be no more events,
	// close channel and wait, which will
	// panic if more events come in
	close(notifyChFW)

	time.Sleep(time.Millisecond * 20)
}

func createTestPool(pingseq map[string][]bool, loadseq []string) (start, stop, notify chan Host) {
	startHostChFW := make(chan Host)
	stopHostChFW := make(chan Host)
	notifyChFW := make(chan Host)

	pingFunc = NewTestPingFunc(pingseq)
	var pool = &Pool{
		Interval:  time.Millisecond,
		FailLimit: 2,
		Receive:   NewTestReceiverFunc(startHostChFW, stopHostChFW),
		Notify:    NewTestNotifyFunc(notifyChFW),
		Load:      NewLoaderFunc(loadseq),
	}

	go pool.Start()

	return startHostChFW, stopHostChFW, notifyChFW
}

// NewTestPingFunc gets a map with a host and a sequence of ping results
// and creates a ping function stub that will yield one by one the result
// sequence for each host, returning 'false' for by default when the sequence
// is over
func NewTestPingFunc(sequence map[string][]bool) func(string) (bool, error) {
	i := make(map[string]int, len(sequence))
	for host := range sequence {
		i[host] = 0 // initialize counter
	}

	var m sync.Mutex
	return func(host string) (bool, error) {
		defer func() {
			m.Unlock()
			recover()
		}()

		m.Lock()
		value := sequence[host][i[host]]
		i[host]++
		return value, nil
	}
}

// NewTestReceiverFunc gets 2 Hosts channel and returns a function that
// forwards whatever is put into those channels into the system injected
// start and stop channels
func NewTestReceiverFunc(startFW, stopFW <-chan Host) Receiver {
	return func(startCh, stopCh chan<- Host) {
		for {
			select {
			case s := <-startFW:
				startCh <- s
			case s := <-stopFW:
				stopCh <- s
			}
		}
	}
}

// NewTestNotifyFunc gets a channel and returns a function that
// forwards whatever is put into the system's notify channel
func NewTestNotifyFunc(notifyFw chan<- Host) Notifier {
	return func(notify <-chan Host) {
		for {
			value := <-notify
			notifyFw <- value
		}
	}
}

// NewLoaderFunc gets the sequence of hostnames/IPs and returns a Loader
// function that when called will insert them as Host structs into the
// start channel at boot time.
func NewLoaderFunc(hosts []string) Loader {
	return func(load chan<- Host) {
		for _, host := range hosts {
			load <- Host{host, false}
		}
	}
}

// Log output catcher
type SkipLog string

func (s SkipLog) Write(p []byte) (n int, err error) {
	return len(p), nil
}
