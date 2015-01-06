package main

import (
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/pinggg/pingd"
	"github.com/pinggg/pingd/io/redis"
	"github.com/pinggg/pingd/ping"
)

// See flags
var (
	redisAddr string
	redisDB   int
	failLimit int
	interval  time.Duration
)

func main() {
	flag.StringVar(&redisAddr, "redis", ":6379", "Redis IP:port")
	flag.IntVar(&redisDB, "redisDB", 0, "Redis DB [0..15]")
	flag.IntVar(&failLimit, "failLimit", 6, "number failed ping attempts in a row to consider host down")
	flag.DurationVar(&interval, "interval", 10*time.Second, "seconds between each ping")
	flag.DurationVar(&ping.TimeOut, "timeOut", 5*time.Second, "seconds for single ping timeout")
	flag.Parse()

	var pool = &pingd.Pool{
		Interval:  interval,
		FailLimit: failLimit,
		Receive:   redis.NewReceiverFunc(redisAddr, redisDB),
		Notify:    redis.NewNotifierFunc(redisAddr, redisDB),
		Load:      redis.NewLoaderFunc(redisAddr, redisDB),
	}

	pool.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c // Exit on interrupt
}
