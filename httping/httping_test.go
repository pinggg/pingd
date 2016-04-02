package httping

import (
	"testing"
	"time"
)

var pingtests = []struct {
	url  string
	ping bool
	err  string
}{
	{"http://httpbin.org/get", true, ""},
	{"https://httpbin.org/get", true, ""},
	{"http://httpbin.org/redirect/1", true, ""},
	{"http://fail.ping.gg", false, "Head http://fail.ping.gg: dial tcp: lookup fail.ping.gg: no such host"},
	{"https://httpbin.org/status/418", false, "418 I'M A TEAPOT"},
	{"https://httpbin.org/redirect/1", true, ""},
	{"http://httpbin.org/delay/3", false, "Head http://httpbin.org/delay/3: net/http: request canceled (Client.Timeout exceeded while awaiting headers)"},
}

func TestPing(t *testing.T) {
	TimeOut = time.Second
	for _, tt := range pingtests {
		ping, err := Ping(tt.url)
		if ping {
			if !tt.ping || err != nil {
				t.Errorf("Incorrect ping for host: %s resulted: %t with error: %s", tt.url, ping, err.Error())
			}
		}

		if !ping {
			if tt.ping || err.Error() != tt.err {
				t.Errorf("Incorrect ping for host: %s resulted: %t with error: %s", tt.url, ping, err.Error())
			}
		}
	}
}
