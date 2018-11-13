package alert

import (
	"sync"
	"time"
	"../storage"
	"../util"
	"../response"
	"github.com/tehmoon/errors"
)

type Manager struct {
	sync *sync.Mutex
	config *ManagerConfig
	triggerManager *TriggerManager
	index map[string]*Alert
	ruleIndex map[string][]*RuleIndex
	ruleIndexSync *sync.Mutex
	configChan chan *AlertConfig
}

type RuleIndex struct {
	From time.Time
	To time.Time
	Value interface{}
}

type ManagerConfig struct {
	PublicURL string
	StorageManager *storage.Manager
	ResponseManager *response.Manager
	Exec string
}

func NewManager(config *ManagerConfig) (manager *Manager) {
	manager = &Manager{
		sync: &sync.Mutex{},
		config: config,
		triggerManager: NewTriggerManager(config.PublicURL),
		index: make(map[string]*Alert),
		configChan: make(chan *AlertConfig, 0),
		ruleIndex: make(map[string][]*RuleIndex),
		ruleIndexSync: &sync.Mutex{},
	}

	for i := 0; i < 5; i++ {
		go func(m *Manager) {
			for {
				select {
					case config := <- m.configChan:
						alert := NewAlert(config)
						m.index[alert.Id] = alert

						tc := &TriggerConfig{
							RuleId: config.Rule.Id(),
							AlertEvery: config.Rule.AlertEvery,
							Alert: alert,
						}

						m.triggerManager.Trigger(tc)

						if config.Rule.Response != nil {
							err := m.config.ResponseManager.Add(&response.ResponseConfig{
								Rule: config.Rule,
								TriggeredAt: alert.TriggeredAt,
								Count: config.Count,
								Value: config.Value,
							})
							if err != nil {
								err = errors.Wrapf(err, "Error triggering response for rule %q", config.Rule.Name())
								util.Println(err.Error())
							}
						}

						err := alert.Save(m.config.StorageManager.Store)
						if err != nil {
							err = errors.Wrapf(err, "Error saving alert id: %q", alert.Id)
							util.Println(err.Error())
						}
				}
			}
		}(manager)
	}

	return manager
}

func (m Manager) Trigger(config *AlertConfig) {
	m.configChan <- config
}
