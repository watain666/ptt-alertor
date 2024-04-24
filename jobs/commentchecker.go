package jobs

import (
	"context"
	"sync"
	"time"

	log "github.com/Ptt-Alertor/logrus"

	"fmt"

	"github.com/watain666/ptt-alertor/models"
	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/ptt/web"
)

var cmtcker *commentChecker
var cmtOnce sync.Once

// commentChecker embedding Checker for checking comment
type commentChecker struct {
	Checker
	Article article.Article
	ch      chan commentChecker
}

// NewCommentChecker return Empty PushChecker pointer
func NewCommentChecker() *commentChecker {
	cmtOnce.Do(func() {
		cmtcker = &commentChecker{}
		cmtcker.duration = 500 * time.Millisecond
		cmtcker.done = make(chan struct{})
		cmtcker.ch = make(chan commentChecker)
	})
	return cmtcker
}

func (cc commentChecker) String() string {
	return fmt.Sprintf("推文@%s\n\n%s\n%s\n%s", cc.Article.Board, cc.Article.Title, cc.Article.Link, cc.Article.Comments.String())
}

func (cc commentChecker) Stop() {
	cc.done <- struct{}{}
	log.Info("Comment Checker Stop")
}

// Run start job
func (cc commentChecker) Run() {
	ach := make(chan article.Article)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				codes := new(article.Articles).List()
				for _, code := range codes {
					time.Sleep(cc.duration)
					go cc.checkComments(code, ach)
				}
			}
		}
	}()

	for {
		select {
		case a := <-ach:
			cc.Article = a
			cc.checkSubscribers()
		case pc := <-cc.ch:
			ckCh <- pc
		case <-cc.done:
			cancel()
			for len(ach) > 0 {
				<-ach
			}
			return
		}
	}
}

func (cc commentChecker) checkComments(code string, ach chan article.Article) {
	a := models.Article().Find(code)
	if a.Board == "" || a.Code == "" {
		return
	}
	new, err := web.FetchArticle(a.Board, a.Code)
	if _, ok := err.(web.URLNotFoundError); ok {
		cc.destroyComments(a)
	}
	if subs, _ := a.Subscribers(); len(subs) == 0 {
		cc.destroyComments(a)
	}
	newComments := make([]article.Comment, 0)
	if new.LastPushDateTime.After(a.LastPushDateTime) {
		for _, comment := range new.Comments {
			if comment.DateTime.After(a.LastPushDateTime) {
				newComments = append(newComments, comment)
			}
		}
		a.LastPushDateTime = new.LastPushDateTime
		a.Comments = newComments
		a.Save()
		log.WithFields(log.Fields{
			"board": a.Board,
			"code":  a.Code,
		}).Info("Updated Comments")
		ach <- a
	}
}

func (cc commentChecker) destroyComments(a article.Article) {
	if err := a.Destroy(); err != nil {
		log.WithError(err).Warning("Destroy Comment Failed")
	}
	log.WithFields(log.Fields{
		"board": a.Board,
		"code":  a.Code,
	}).Info("Destroy Comments")
}

func (cc commentChecker) checkSubscribers() {
	subs, err := cc.Article.Subscribers()
	if err != nil {
		log.WithError(err).Error("Get Subscribers Failed")
	}

	for _, account := range subs {
		go cc.send(account)
	}
}

func (cc commentChecker) send(account string) {
	cc.board = cc.Article.Board
	cc.subType = "push"
	cc.word = cc.Article.Code
	cc.Profile = models.User().Find(account).Profile
	cc.ch <- cc
}
