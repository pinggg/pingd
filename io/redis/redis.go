package redis

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"fmt"

	"github.com/pinggg/pingd"
)

var (
	HostListRK = "hostlist" // Redis key for host list
	StartRK    = "start"    // Redis pubsub key for start
	StopRK     = "stop"     // Redis pubsub key for stop
	UpRK       = "up"       // Redis pubsub key for up notification
	DownRK     = "down"     // Redis pubsub key for down notification
)

const (
	upStatus   = "up"   // Status value for host up
	downStatus = "down" // Status value for host down

	// when receiving host on the start channel
	// they can be requested to start as "down"
	// adding it at the end eg "example.com down"
	downSuffix = " down"
)

// NewReceiverFunc returns the function that
// listens of redis for start/stop commands
func NewReceiverFunc(redisAddr string, redisDB int) pingd.Receiver {
	return func(startHostCh, stopHostCh chan<- pingd.HostStatus) {
		conPubSub, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			log.Panicln(err)
		}

		connKV, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			log.Panicln(err)
		}

		servername, _ := os.Hostname()
		conPubSub.Do("CLIENT", "SETNAME", "receive-"+servername)
		conPubSub.Do("SELECT", redisDB)
		connKV.Do("CLIENT", "SETNAME", "receive-"+servername)
		connKV.Do("SELECT", redisDB)

		psc := redis.PubSubConn{conPubSub}
		psc.Subscribe(StartRK, StopRK)

		for {
			switch n := psc.Receive().(type) {
			case redis.Message:
				if n.Channel == StartRK {
					host := string(n.Data)
					down := false
					if strings.HasSuffix(host, downSuffix) {
						down = true
						host = strings.Replace(host, downSuffix, "", 1)
					}

					// Add to the list of pinged hosts
					_, err := connKV.Do("SADD", HostListRK, host)
					if err != nil {
						log.Panicln(err)
					}
					startHostCh <- pingd.HostStatus{Host: host, Down: down}

				} else if n.Channel == StopRK {
					host := string(n.Data)

					// Remove from the list of pinged hosts
					_, err := connKV.Do("SREM", HostListRK, host)
					if err != nil {
						log.Panicln(err)
					}
					stopHostCh <- pingd.HostStatus{Host: host}
				}

			case redis.PMessage:
			case redis.Subscription:
				log.Println("BOOT Listening to " + n.Channel)
			case error:
				log.Printf("error: %v\n", n)
				return
			}
		}
	}
}

// NewNotifierFunc returns the function that
// publishes on redis the up/down events
func NewNotifierFunc(redisAddr string, redisDB int) pingd.Notifier {
	return func(notifyCh <-chan pingd.HostStatus) {
		conn, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			log.Panicln(err)
		}

		servername, _ := os.Hostname()
		_, err = conn.Do("CLIENT", "SETNAME", "notify-"+servername)
		if err != nil {
			log.Panicln(err)
		}

		_, err = conn.Do("SELECT", redisDB)
		if err != nil {
			log.Panicln(err)
		}

		var h pingd.HostStatus
		for {
			select {
			case h = <-notifyCh:
				switch h.Down {
				// DOWN
				case true:
					log.Println("DOWN " + h.Host)
					conn.Send("PUBLISH", DownRK, fmt.Sprintf("%s %s", h.Host, h.Reason))
					conn.Send("SET", "status-"+h.Host, downStatus)
					conn.Flush()
					// UP
				case false:
					log.Println("UP " + h.Host)
					conn.Send("PUBLISH", UpRK, h.Host)
					conn.Send("SET", "status-"+h.Host, upStatus)
					conn.Flush()
				}
			}
		}
	}
}

// NewLoaderFunc returns the function that loads back
// hosts and last statuses from REDIS in case of reboot
// send them to the startHostCh channel
func NewLoaderFunc(redisAddr string, redisDB int) pingd.Loader {
	return func(startHostCh chan<- pingd.HostStatus) {
		log.Println("BOOT Loading hosts")
		conn, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			log.Panicln(err)
		}

		servername, _ := os.Hostname()
		_, err = conn.Do("CLIENT", "SETNAME", "load-"+servername)
		if err != nil {
			log.Panicln(err)
		}
		_, err = conn.Do("SELECT", redisDB)
		if err != nil {
			log.Panicln(err)
		}

		hosts, err := redis.Strings(conn.Do("SMEMBERS", HostListRK))
		if err != nil {
			log.Panicln(err)
		}

		for _, host := range hosts {
			var down bool

			// Check for status
			status, err := redis.String(conn.Do("GET", "status-"+host))
			if err != nil {
				log.Println("ERROR loading status of " + host + ". Assuming UP")
			}
			if status == downStatus {
				down = true
			}

			// load into process
			startHostCh <- pingd.HostStatus{Host: host, Down: down}

			// slow a bit loading process
			time.Sleep(time.Millisecond * 10)
		}

		log.Println("BOOT " + strconv.Itoa(len(hosts)) + " hosts loaded")
		log.Println("BOOT Ready")
	}

}
