#### Disclaimer: If you are new to Go this is not a good place to learn best practices, the code is not very idiomatic and there's probably a few bad ideas.

### What's this?

pingd is the ping engine used by the [ping.gg monitoring service](https://ping.gg). It allows to simultaneously ping hundreds of IPs and can manage which hosts are being pinged during runtime.

The system is meant to be composed (or embedded) by you and requires up to 3 compatible functions to:

1. load an initial set of hosts to monitor
2. receive commands to start and stop monitoring a given host
3. publish host UP or host DOWN events

These functions must be the following types respectively.

```go
type Loader func(chan<- Host)
type Receiver func(chan<- Host, chan<- Host)
type Notifier func(<-chan Host)
```

There are some implementations of these functions available under pingd/io.

### Usage example

NOTE: Before you run anything, remember that ICMP echo (ping) requires root privileges for raw socket access.
If you don't want to sudo every time consider using [setuid](http://www.cyberciti.biz/faq/unix-bsd-linux-setuid-file/)

To create your own private ping.gg alternative:

```bash
 # get the example
export GOPATH=$PWD
go get github.com/pinggg/pingd/examples/httpmail

 # start pinging 8.8.8.8 and 8.8.4.4 and email you if they go down
sudo bin/httpmail -email=mymail@example.org 8.8.8.8 8.8.4.4

 # add a new host while running
curl localhost:7700/4.4.2.2

 # stop pinging 8.8.4.4
curl -XDELETE localhost:7700/8.8.4.4
```

Keep in mind that this is just an example which assumes that there is a local mail server running on port 25. Take a look at the [sendMail function](https://github.com/pinggg/pingd/blob/master/examples/httpmail/cmd.go#L55) and adapt it to your needs. For example, to [send the emails via Gmail](https://github.com/jordan-wright/email#sending-email-using-gmail).

https://ping.gg uses in production a configuration like the [redis example](https://github.com/pinggg/pingd/blob/master/examples/redis/cmd.go) allowing the website to interact with pingd via redis pub/sub.

You can add your own functions to have pingd interact with the world. For example, switching on some red light with the help of a Raspberry Pi.
