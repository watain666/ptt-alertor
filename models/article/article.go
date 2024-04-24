package article

import (
	"regexp"
	"strconv"
	"strings"

	"time"

	"fmt"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/models/pushsum"
	"github.com/watain666/ptt-alertor/myutil"
)

const prefix = "article:"
const subsSuffix = ":subs"

type Article struct {
	ID               int    `json:"ID,omitempty"`
	Code             string `json:"code,omitempty"`
	Title            string
	Link             string
	Date             string    `json:"Date,omitempty"`
	Author           string    `json:"Author,omitempty"`
	Comments         Comments  `json:"comments,omitempty"`
	LastPushDateTime time.Time `json:"lastPushDateTime,omitempty"`
	Board            string    `json:"board,omitempty"`
	PushSum          int       `json:"pushSum,omitempty"`
	drive            Driver
}

type Driver interface {
	Find(code string, article *Article)
	Save(a Article) error
	Delete(code string) error
}

func NewArticle(drive Driver) *Article {
	return &Article{
		drive: drive,
	}
}

func (a Article) ParseID(Link string) (id int) {
	reg, err := regexp.Compile("https?://www.ptt.cc/bbs/.*/[GM]\\.(\\d+)\\..*")
	if err != nil {
		log.Fatal(err)
	}
	strs := reg.FindStringSubmatch(Link)
	if len(strs) < 2 {
		return 0
	}
	id, err = strconv.Atoi(strs[1])
	if err != nil {
		return 0
	}
	return id
}

func (a Article) MatchKeyword(keyword string) bool {
	if strings.Contains(keyword, "&") {
		keywords := strings.Split(keyword, "&")
		for _, keyword := range keywords {
			if !matchKeyword(a.Title, keyword) {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(keyword, "regexp:") {
		return matchRegex(a.Title, keyword)
	}
	return matchKeyword(a.Title, keyword)
}

// Exist check article exist or not
func (a Article) Exist() (bool, error) {
	conn := connections.Redis()
	defer conn.Close()

	bl, err := redis.Bool(conn.Do("EXISTS", prefix+a.Code+subsSuffix, "board"))
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return bl, err
}

func (a Article) Find(code string) Article {
	a.drive.Find(code, &a)
	return a
}

func (a Article) Save() error {
	return a.drive.Save(a)
}

func (a Article) Destroy() error {
	if err := a.drive.Delete(a.Code); err != nil {
		return err
	}

	conn := connections.Redis()
	defer conn.Close()

	_, err := conn.Do("DEL", prefix+a.Code+subsSuffix)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) AddSubscriber(account string) error {
	conn := connections.Redis()
	defer conn.Close()

	_, err := conn.Do("SADD", prefix+a.Code+subsSuffix, account)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) Subscribers() ([]string, error) {
	conn := connections.Redis()
	defer conn.Close()

	accounts, err := redis.Strings(conn.Do("SMEMBERS", prefix+a.Code+subsSuffix))
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return accounts, err
}

func (a Article) RemoveSubscriber(sub string) error {
	conn := connections.Redis()
	defer conn.Close()

	_, err := conn.Do("SREM", prefix+a.Code+subsSuffix, sub)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}

func (a Article) String() string {
	return a.Title + "\r\n" + a.Link
}

func (a Article) StringWithPushSum() string {
	sumStr := strconv.Itoa(a.PushSum)
	if text, ok := pushsum.NumTextMap[a.PushSum]; ok {
		sumStr = text
	}
	return fmt.Sprintf("%s %s\r\n%s", sumStr, a.Title, a.Link)
}

func matchRegex(title string, regex string) bool {
	pattern := strings.TrimPrefix(regex, "regexp:")
	b, err := regexp.MatchString(pattern, title)
	if err != nil {
		return false
	}
	return b
}

func matchKeyword(title string, keyword string) bool {
	if strings.HasPrefix(keyword, "!") {
		excludeKeyword := strings.Trim(keyword, "!")
		return !containKeyword(title, excludeKeyword)
	}
	return containKeyword(title, keyword)
}

func containKeyword(title string, keyword string) bool {
	return strings.Contains(strings.ToLower(title), strings.ToLower(keyword))
}
