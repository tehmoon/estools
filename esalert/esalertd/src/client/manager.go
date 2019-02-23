package client

import (
	"github.com/olivere/elastic"
	"sync"
	"github.com/tehmoon/errors"
)

type Manager struct {
	sync *sync.Mutex
	client *elastic.Client
}

type ManagerConfig struct {
	ElasticConfigs []elastic.ClientOptionFunc
}

func NewManager(config *ManagerConfig) (manager *Manager, err error) {
	client, err := elastic.NewClient(config.ElasticConfigs...)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new elastic client")
	}

	manager = &Manager{
		sync: &sync.Mutex{},
		client: client,
	}

	return manager, nil
}

func (m Manager) Search(index ...string) (service *elastic.SearchService) {
	return m.client.Search(index...)
}

func (m Manager) ClearScroll() (service *elastic.ClearScrollService) {
	return m.client.ClearScroll()
}

func (m Manager) Scroll(index ...string) (service *elastic.ScrollService) {
	return m.client.Scroll(index...)
}
