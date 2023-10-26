package mailer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	gomail "gopkg.in/mail.v2"
)

var (
	runMu sync.Mutex
	// stores if a daemon is running
	// prevents running multiple mailers simultaneously
	running uint32
)

// TextMailer implements Mailer interface
// it runs a daemon waiting for text messages to send to a predefined address
type TextMailer struct {
	cfg      Config
	messages chan *gomail.Message
	running  bool
}

// New returns a Mailer implementation
func New(cfg Config) *TextMailer {
	return &TextMailer{
		cfg: cfg,
	}
}

// Run starts the mailer daemon
func (mailer *TextMailer) Run(stop <-chan struct{}) error {
	if atomic.LoadUint32(&running) == 1 {
		return newAlreadyRunningError()
	}

	runMu.Lock()

	if running == 1 {
		runMu.Unlock()
		return newAlreadyRunningError()
	}

	// create fresh channel
	mailer.messages = make(chan *gomail.Message)
	go mailer.daemon(stop)
	// set running state true
	atomic.StoreUint32(&running, 1)
	mailer.running = true

	runMu.Unlock()

	return nil
}

// SendMessage prepares new messages and sends them
// Caller is responsible for proper escaping of message in case of e.g. HTML
func (mailer *TextMailer) SendMessage(contentType, messageText string) error {
	if !mailer.running {
		return newNotRunningError()
	}

	// prepare message
	msg := gomail.NewMessage()
	msg.SetHeader("From", mailer.cfg.From)
	msg.SetHeader("To", mailer.cfg.To)
	msg.SetHeader("Subject", mailer.cfg.Subject)
	msg.SetBody(contentType, messageText)

	mailer.messages <- msg
	return nil
}

// daemon listens for messages on the channel and sends them
func (mailer *TextMailer) daemon(stop <-chan struct{}) {
	// prepare smpt dialer
	dialer := gomail.NewDialer(mailer.cfg.Server, mailer.cfg.Port, mailer.cfg.User, mailer.cfg.Password)

	var s gomail.SendCloser
	var err error
	// dialer status: is open or closed
	open := false
	for {
		select {
		case msg := <-mailer.messages:
			if !open {
				if s, err = dialer.Dial(); err != nil {
					log.Error().
						Err(err).
						Msg("could not dial smtp server")
				}
				open = true
			}
			if err := gomail.Send(s, msg); err != nil {
				log.Error().
					Err(err).
					Msg("could not send mail")
			}
			// Close the connection to the SMTP server if no email was sent in
			// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := s.Close(); err != nil {
					log.Error().
						Err(err).
						Msg("could not close sender")
				}
				open = false
			}
		case <-stop:
			runMu.Lock()

			// set running state false
			atomic.StoreUint32(&running, 0)
			mailer.running = false

			runMu.Unlock()
			return
		}

	}
}
