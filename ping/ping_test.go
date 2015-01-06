package ping

import (
	"testing"
	"time"
)

var pingtests = []struct {
	host string
	ping bool
	err  string
}{
	{"127.0.0.1", true, ""},
	{"8.8.8.8", true, ""},
	{"google.com", true, ""},
	{"128.0.0.1", false, "read ip 128.0.0.1: i/o timeout"},
	{"fail.ping.gg", false, "dial ip:icmp: lookup fail.ping.gg: no such host"},
}

func TestPing(t *testing.T) {

	TimeOut = time.Second / 2

	for _, tt := range pingtests {
		ping, err := Ping(tt.host)
		if ping {
			if !tt.ping || err != nil {
				t.Errorf("Incorrect ping for host: %s resulted: %t with error: %s", tt.host, ping, err.Error())
			}
		}

		if !ping {
			if tt.ping || err.Error() != tt.err {
				t.Errorf("Incorrect ping for host: %s resulted: %t with error: %s", tt.host, ping, err.Error())
			}
		}
	}
}
