package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"strings"

	"strconv"

	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/models"
	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/models/pushsum"
	"github.com/watain666/ptt-alertor/models/subscription"
	"github.com/watain666/ptt-alertor/models/user"
	"github.com/watain666/ptt-alertor/ptt/web"
)

// NewPushSumKeyReplacer Job schedule must longer than overduehour
const overdueHour = 48 * time.Hour

var psCker *pushSumChecker
var pscOnce sync.Once

type pushSumChecker struct {
	Checker
	ch       chan pushSumChecker
	duration time.Duration
}

func NewPushSumChecker() *pushSumChecker {
	pscOnce.Do(func() {
		psCker = &pushSumChecker{
			duration: 500 * time.Millisecond,
		}
		psCker.done = make(chan struct{})
		psCker.ch = make(chan pushSumChecker)
	})
	return psCker
}

func (psc pushSumChecker) String() string {
	textMap := map[string]string{
		"pushup":   "推文數",
		"pushdown": "噓文數",
	}
	subType := textMap[psc.subType]
	return fmt.Sprintf("%s@%s\r\n看板：%s；%s：%s%s", psc.word, psc.board, psc.board, subType, psc.word, psc.articles.StringWithPushSum())
}

type BoardArticles struct {
	board    string
	articles article.Articles
}

func (psc pushSumChecker) Stop() {
	psc.done <- struct{}{}
	log.Info("Pushsum Checker Stop")
}

func (psc pushSumChecker) Run() {
	baCh := make(chan BoardArticles)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				boards := pushsum.List()
				for _, board := range boards {
					ba := BoardArticles{board: board}
					time.Sleep(psc.duration)
					go psc.crawlArticles(ba, baCh)
				}
			}
		}
	}()

	for {
		select {
		case ba := <-baCh:
			psc.board = ba.board
			if len(ba.articles) > 0 {
				go psc.checkSubscribers(ba)
			}
		case pscker := <-psc.ch:
			ckCh <- pscker
		case <-psc.done:
			cancel()
			for len(baCh) > 0 {
				<-baCh
			}
			for len(psc.ch) > 0 {
				<-psc.ch
			}
			return
		}
	}
}

func (psc pushSumChecker) crawlArticles(ba BoardArticles, baCh chan BoardArticles) {
	currentPage, err := web.CurrentPage(ba.board)
	if err != nil {
		log.WithFields(log.Fields{
			"board": ba.board,
		}).WithError(err).Error("Get CurrentPage Failed")
		baCh <- ba
		return
	}

Page:
	for page := currentPage; page > 0; page-- {
		articles, _ := web.FetchArticles(ba.board, page)
		for i := len(articles) - 1; i > 0; i-- {
			a := articles[i]
			if a.ID == 0 {
				continue
			}
			loc := time.FixedZone("CST", 8*60*60)
			t, err := time.ParseInLocation("1/02", a.Date, loc)
			now := time.Now()
			nowDate := now.Truncate(24 * time.Hour)
			if t.Month() > now.Month() {
				t = t.AddDate(now.Year()-1, 0, 0)
			} else {
				t = t.AddDate(now.Year(), 0, 0)
			}
			if err != nil {
				log.WithFields(log.Fields{
					"board": ba.board,
					"page":  page,
				}).WithError(err).Error("Parse DateTime Error")
				continue
			}
			if nowDate.After(t.Add(overdueHour)) {
				break Page
			}
			ba.articles = append(ba.articles, a)
		}
	}

	log.WithFields(log.Fields{
		"board": ba.board,
		"total": len(ba.articles),
	}).Info("PushSum Crawl Finish")

	baCh <- ba
}

func (psc pushSumChecker) checkSubscribers(ba BoardArticles) {
	subs := pushsum.ListSubscribers(ba.board)
	for _, account := range subs {
		u := models.User().Find(account)
		psc.Profile = u.Profile
		go psc.checkPushSum(u, ba, checkUp)
		go psc.checkPushSum(u, ba, checkDown)
	}
}

type checkPushSumFn func(*pushSumChecker, subscription.Subscription, article.Articles) (article.Articles, []int)

func checkUp(psc *pushSumChecker, sub subscription.Subscription, articles article.Articles) (upArticles article.Articles, ids []int) {
	psc.word = strconv.Itoa(sub.Up)
	psc.subType = "pushup"
	if sub.Up != 0 {
		for _, a := range articles {
			if a.PushSum >= sub.Up {
				upArticles = append(upArticles, a)
				ids = append(ids, a.ID)
			}
		}
	}
	return upArticles, ids
}

func checkDown(psc *pushSumChecker, sub subscription.Subscription, articles article.Articles) (downArticles article.Articles, ids []int) {
	down := sub.Down * -1
	psc.word = strconv.Itoa(down)
	psc.subType = "pushdown"
	if sub.Down != 0 {
		for _, a := range articles {
			if a.PushSum <= down {
				downArticles = append(downArticles, a)
				ids = append(ids, a.ID)
			}
		}
	}
	return downArticles, ids
}

func (psc pushSumChecker) checkPushSum(u user.User, ba BoardArticles, checkFn checkPushSumFn) {
	var articles article.Articles
	var ids []int
	for _, sub := range u.Subscribes {
		if strings.EqualFold(sub.Board, ba.board) {
			articles, ids = checkFn(&psc, sub, ba.articles)
		}
	}
	if len(articles) > 0 {
		psc.articles = psc.toSendArticles(ids, articles)
		if len(psc.articles) > 0 {
			psc.ch <- psc
		}
	}
}

func (psc pushSumChecker) toSendArticles(ids []int, articles article.Articles) article.Articles {
	kindMap := map[string]string{
		"pushup":   "up",
		"pushdown": "down",
	}
	ids = pushsum.DiffList(psc.Profile.Account, psc.board, kindMap[psc.subType], ids...)
	diffIds := make(map[int]bool)
	for _, id := range ids {
		diffIds[id] = true
	}
	sendArticles := make(article.Articles, 0)
	for _, a := range articles {
		if diffIds[a.ID] {
			sendArticles = append(sendArticles, a)
		}
	}
	return sendArticles
}
