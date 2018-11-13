package alert

import (
	"sync"
	"time"
)

type TriggerManager struct {
	alertTriggers map[string]time.Time
	configChan chan *TriggerConfig
	sync *sync.Mutex
	publicURL string
}

type TriggerConfig struct {
	RuleId string
	AlertEvery time.Duration
	Alert *Alert
}

func (tm *TriggerManager) Trigger(tc *TriggerConfig) {
	tm.AddAndFlush(tc)
}

func (tm *TriggerManager) AddAndFlush(config *TriggerConfig) {
	tm.sync.Lock()
	tm.addAndFlush(config)
	tm.sync.Unlock()
}

func (tm *TriggerManager) addAndFlush(config *TriggerConfig) {
	triggeredAt, found := tm.alertTriggers[config.RuleId]
	if ! found {
		tm.alertTriggers[config.RuleId] = config.Alert.TriggeredAt
		config.Alert.Trigger(true, tm.publicURL)

		return
	}

	if triggeredAt.Add(config.AlertEvery).Unix() <= config.Alert.TriggeredAt.Unix() {
		tm.alertTriggers[config.RuleId] = config.Alert.TriggeredAt
		config.Alert.Trigger(true, tm.publicURL)

		return
	}

	config.Alert.Trigger(false, tm.publicURL)
}

// The triggermanager will check if when the alert is triggered, if it needs
// to alert people. It does so by setting the `alert` field to true.
// Alert plugins should respect that field and send the alert only when
// necessary. Otherwise, they are free to log it somewhere and or implement
// additional logic.
func NewTriggerManager(publicURL string) (tm *TriggerManager) {
	tm = &TriggerManager{
		sync: &sync.Mutex{},
		alertTriggers: make(map[string]time.Time),
		publicURL: publicURL,
		configChan: make(chan *TriggerConfig, 0),
	}

	return tm
}
