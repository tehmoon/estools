package rules

import (
	"sync"
	"github.com/tehmoon/errors"
)

type Manager struct {
	config *ManagerConfig
	sync *sync.RWMutex
	rules []*Rule
}

type ManagerConfig struct {
	Owners []string
	Exec string
	Index string
}

func NewManager(config *ManagerConfig) (*Manager) {
	return &Manager{
		config: config,
		sync: &sync.RWMutex{},
	}
}

func (m *Manager) LoadRules(p string) (err error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	rules, err := loadRules(p, m.config.Index, m.config.Exec, m.config.Owners)
	if err != nil {
		return errors.Wrap(err, "Error loading rules")
	}

	m.rules = rules

	return nil
}

func (m Manager) Rules() (rules []*Rule) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	return m.rules
}
