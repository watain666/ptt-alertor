package shorturl

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/garyburd/redigo/redis"

	"strconv"

	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

const redisPrefix = "sum:"

var url = os.Getenv("APP_HOST") + "/redirect/"

func Gen(longURL string) string {
	data := []byte(longURL)
	sum := fmt.Sprintf("%x", md5.Sum(data))
	sum += strconv.FormatInt(time.Now().Unix(), 10)
	conn := connections.Redis()
	defer conn.Close()
	_, err := conn.Do("SET", redisPrefix+sum, longURL, "EX", 600)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	shortURL := url + sum
	return shortURL
}

func Original(sum string) string {
	conn := connections.Redis()
	defer conn.Close()
	key := redisPrefix + sum
	u, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return u
}
