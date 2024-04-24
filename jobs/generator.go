package jobs

import (
	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/models"
	"github.com/watain666/ptt-alertor/models/author"
	"github.com/watain666/ptt-alertor/models/keyword"
	"github.com/watain666/ptt-alertor/models/pushsum"
	"github.com/watain666/ptt-alertor/models/subscription"
)

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (gb Generator) Run() {
	boardNameBool := make(map[string]bool)

	for _, bd := range models.Board().All() {
		boardNameBool[bd.Name] = true
	}

	for _, u := range models.User().All() {
		for _, sub := range u.Subscribes {
			if !boardNameBool[sub.Board] {
				addBoard(sub.Board)
			}
			if sub.PushSum != subscription.EmptyPushSum {
				addPushsumSub(u.Profile.Account, sub.Board)
			}
			if len(sub.Keywords) > 0 {
				addKeywordSub(u.Profile.Account, sub.Board)
			}
			if len(sub.Authors) > 0 {
				addAuthorSub(u.Profile.Account, sub.Board)
			}
			if len(sub.Articles) > 0 {
				for _, a := range sub.Articles {
					addArticleSub(u.Profile.Account, a)
				}
			}
		}
	}
	log.Info("Generated Done")
}

func addBoard(boardName string) {
	bd := models.Board()
	bd.Name = boardName
	bd.Create()
	log.WithField("board", bd.Name).Info("Added Board")
}

func addPushsumSub(account, board string) {
	pushsum.Add(board)
	pushsum.AddSubscriber(board, account)
	log.WithField("board", board).Info("Added PushSum Board and Subscriber")
}

func addKeywordSub(account, board string) {
	keyword.AddSubscriber(board, account)
	log.WithField("board", board).Info("Added Keyword Subscriber")
}

func addAuthorSub(account, board string) {
	author.AddSubscriber(board, account)
	log.WithField("board", board).Info("Added Author Subscriber")
}

func addArticleSub(account, articleID string) {
	a := models.Article()
	a.Code = articleID
	a.AddSubscriber(account)
	log.WithField("article", articleID).Info("Added Article Subscriber")
}
