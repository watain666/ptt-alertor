package mail

import (
	"os"
	"strconv"

	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/models/article"
	"gopkg.in/mailgun/mailgun-go.v1"
)

type Mail struct {
	Title
	Body
	Receiver string
}

type Title struct {
	BoardName       string
	Keyword         string
	articleQuantity int
}

type Body struct {
	Articles article.Articles
}

func (title Title) String() string {
	return "[PttAlertor] 在 " + title.BoardName + " 板有 " + strconv.Itoa(title.articleQuantity) + " 篇關於「" + title.Keyword + "」的文章發表"
}

func (body Body) String() string {
	return body.Articles.String() + "\r\n\r\nSend From Ptt Alertor"
}

func (mail Mail) Send() {
	mg := newMailgun()

	mail.articleQuantity = len(mail.Body.Articles)
	message := mailgun.NewMessage(
		"PttAlertor@mg.dinolai.com",
		mail.Title.String(),
		mail.Body.String(),
		mail.Receiver)
	resp, id, err := mg.Send(message)
	if err != nil {
		log.WithError(err).Error("Sent Email Failed")
	} else {
		log.WithFields(log.Fields{
			"ID":   id,
			"Resp": resp,
		}).Info("Sent Email")
	}
}

var (
	domain       = os.Getenv("MAILGUN_DOMAIN")
	apiKey       = os.Getenv("MAILGUN_APIKEY")
	publicAPIKey = os.Getenv("MAILGUN_PUBLIC_APIKEY")
)

func newMailgun() mailgun.Mailgun {
	return mailgun.NewMailgun(domain, apiKey, publicAPIKey)
}
