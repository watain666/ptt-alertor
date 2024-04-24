package counter

import (
	log "github.com/Ptt-Alertor/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/myutil"
)

func Alert() (int, error) {
	conn := connections.Redis()
	defer conn.Close()
	count, err := redis.Int(conn.Do("GET", "counter:alert"))
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return count, err
}

func IncrAlert() error {
	conn := connections.Redis()
	defer conn.Close()
	count, err := redis.Int(conn.Do("INCR", "counter:alert"))
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	publishAlert(count)
	return err
}

func publishAlert(count int) error {
	conn := connections.Redis()
	defer conn.Close()
	_, err := conn.Do("PUBLISH", "alert-counter", count)
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error()
	}
	return err
}
