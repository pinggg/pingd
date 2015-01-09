package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jordan-wright/email"

	"github.com/pinggg/pingd"
	"github.com/pinggg/pingd/io/http"
	"github.com/pinggg/pingd/io/mail"
	"github.com/pinggg/pingd/io/std"
	"github.com/pinggg/pingd/ping"
)

// See flags
var (
	emailAddr  string
	listenAddr string

	interval  time.Duration
	failLimit int
)

func main() {
	flag.StringVar(&emailAddr, "email", "me@example.org", "email recipient for notificiations")
	flag.StringVar(&listenAddr, "listen", ":7700", "webserver listen address")
	flag.IntVar(&failLimit, "failLimit", 4, "number failed ping attempts in a row to consider host down")
	flag.DurationVar(&interval, "interval", 5*time.Second, "seconds between each ping")
	flag.DurationVar(&ping.TimeOut, "timeOut", 5*time.Second, "seconds for single ping timeout")
	flag.Parse()

	// read non flag arguments as hosts to start monitoring
	hosts := flag.Args()

	var pool = &pingd.Pool{
		Interval:  interval,
		FailLimit: failLimit,
		Receive:   http.NewReceiverFunc(listenAddr),          // start/stop commands via HTTP
		Notify:    mail.NewNotifierFunc(emailAddr, sendMail), // notify up/down via email
		Load:      std.NewLoaderFunc(hosts),                  // load initial hosts from command line
	}

	pool.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c // Exit on interrupt
}

// Replace this localhost version with your appropriate mail function
// https://github.com/jordan-wright/email#email
func sendMail(recipient, message string) {
	e := email.NewEmail()

	// Change From to something more appropriate
	e.From = recipient
	e.To = []string{recipient}
	e.Subject = message

	err := e.Send("localhost:25", nil)
	if err != nil {
		log.Fatal(err)
	}
}
