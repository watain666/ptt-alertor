package jobs

import (
	log "github.com/Ptt-Alertor/logrus"
	"github.com/watain666/ptt-alertor/models/pushsum"
)

type PushSumKeyReplacer struct{}

func NewPushSumKeyReplacer() *PushSumKeyReplacer {
	return &PushSumKeyReplacer{}
}

func (r PushSumKeyReplacer) Run() {
	if err := pushsum.ReplaceBenchKeys(); err != nil {
		log.WithError(err).Error("Replace Pushsum Key Failed")
	}
	log.Info("Replace Pushsum Key Done")
}
