package mail

import (
	"strings"

	"github.com/DusanKasan/parsemail"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	Subject string
	Body    string
}

func ParseBody(email string) Message {
	log.Debug("Parsing body")

	p := strings.NewReader(email)
	emailOut, err := parsemail.Parse(p)
	if err != nil {
		panic(err)
	}

	message := &Message{
		Subject: emailOut.Subject,
		Body:    emailOut.TextBody,
	}

	return *message
}
