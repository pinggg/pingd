package mail

import (
	"fmt"
	"log"

	"github.com/pinggg/pingd"
)

// Mailer is a function which takes a email
// address and a message and sends an email
type Mailer func(recepient string, message string)

// NewNotifierFunc takes a email address and a email sending function
// and will send emails with every up and down event.
func NewNotifierFunc(recepient string, mailerFunc Mailer) pingd.Notifier {
	return func(notify <-chan pingd.Host) {
		for {
			host := <-notify
			status := "UP"
			if host.Down {
				status = "DOWN"
			}
			message := fmt.Sprintf("host %s is %s", host.Host, status)

			mailerFunc(recepient, message)
			log.Printf(message)
		}
	}
}
