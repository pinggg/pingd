#### Disclaimer: If you are new to Go this is not a good place to learn best practices, the code is not very idiomatic and there's probably a few bad ideas.

### What's this?

pingd is the ping engine used by the [ping.gg monitoring service](https://ping.gg). It allows to simultaneously ping hundreds of IPs and can manage which hosts are being pinged during runtime.

The system is meant to be composed by you and requires up to 3 compatible functions to:

1. load an initial set of hosts to monitor
2. receive commands to start and stop monitoring a given host
3. publish host UP or host DOWN events

This function must be the types following types respectively.

```go
type Loader func(chan<- Host)
type Receiver func(chan<- Host, chan<- Host)
type Notifier func(<-chan Host)
```

There are some implementations of this functions available under pingd/io.

### Usage example

NOTE: Before you run anything, remember that ICMP echo (ping) requires root privileges for raw socket access.
If you don't want to sudo every time consider setuid http://www.cyberciti.biz/faq/unix-bsd-linux-setuid-file/

To create your own private ping.gg alternative

```bash
 # get the example
export GOPATH=$PWD
go get github.com/pinggg/pingd/examples/httpmail

 # start pinging 8.8.8.8 and 8.8.4.4 and email you if they go down
sudo bin/httpmain -email=mymail@example.org 8.8.8.8 8.8.4.4

 # add an new host while running
curl localhost:7700/4.4.2.2

 # stop pinging another one
curl -XDELETE localhost:7700/8.8.4.4
```

the ping.gg website uses a configuration very close to the redis example to interact via redis pub/sub with pingd.