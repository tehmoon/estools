package storage

import (
	"io"
	"context"
	"github.com/tehmoon/errors"
	"encoding/json"
	"github.com/olivere/elastic"
	"../client"
	"../util"
	"../rules"
	"bytes"
	"compress/gzip"
)

// TODO: pass flags to the storage configuration
type Manager struct {
	storage Storage
	config *ManagerConfig
	work chan *WorkConfig
	cancel chan struct{}
}

type Storage interface {
	Init() (err error)
	Store(id string, data []byte) (err error)
	Get(id string) (data []byte, err error)
}

func (m Manager) Store(config *WorkConfig) (err error) {
	if config.Rule.Type != rules.RuleTypeCount {
		return nil
	}

	m.work <- config

	return nil
}

func (m Manager) Get(id string, writer io.Writer) (err error) {
	data, err := m.storage.Get(id)
	if err != nil {
		return errors.Wrap(err, "Error getting data from storage")
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "Error creating new gzip reader")
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return errors.Wrap(err, "Error reading gzip data from storage")
	}

	return nil
}

type ManagerConfig struct {
	StorageName string
	ClientManager *client.Manager
}

func NewManager(config *ManagerConfig) (manager *Manager, err error) {
	workers := 5
	manager = &Manager{
		config: config,
		work: make(chan *WorkConfig, 0),
		cancel: make(chan struct{}, 0),
	}

	switch config.StorageName {
		case "memory":
			manager.storage = NewMemoryStorage()
		default:
			return nil, errors.Errorf("Storage %q is not supported", config.StorageName)
	}

	err = manager.storage.Init()
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing storage")
	}

	for i := 0; i < workers; i++ {
		go worker(manager.storage, manager.config.ClientManager, manager.work, manager.cancel)
	}

	return manager, nil
}

func worker(storage Storage, cm *client.Manager, work chan *WorkConfig, cancel chan struct{}) {
	buff := bytes.NewBuffer(nil)

	LOOP: for {
		select {
			case job := <- work:
				err := scroll(cm, buff, job)
				if err != nil {
					err = errors.Wrapf(err, "Error scrolling through query for alert id %q", job.Id)
					util.Println(err.Error())
					buff.Reset()
					continue
				}

				data := make([]byte, buff.Len())
				copy(data, buff.Bytes())
				buff.Reset()

				err = storage.Store(job.Id, data)
				if err != nil {
					err = errors.Wrapf(err, "Error storing data for alert id %q", job.Id)
					util.Println(err.Error())
					continue
				}
			case <- cancel:
				break LOOP
		}
	}
}

// TODO: Add data export type, not everything needs to be a scroll
type WorkConfig struct {
	Query elastic.Query
	Rule *rules.Rule
	Id string
}

func scroll(cm *client.Manager, w io.Writer, config *WorkConfig) (err error) {
	// TODO: make variable
	size := 1000
	timestamp := "@timestamp"
	asc := true
	scrollTime := "10s"

	writer := gzip.NewWriter(w)
	defer writer.Close()

	s := cm.Scroll(config.Rule.Index).
		Size(size).
		Sort(timestamp, asc).
		Query(config.Query).
		Scroll(scrollTime)

	res, err := s.Do(context.Background())
	if err != nil {
		if err == io.EOF {
			return nil
		}

		return errors.Wrap(err, "Error calling initial scroll")
	}

	err = processHits(writer, res.Hits.Hits, config)
	if err != nil {
		return errors.Wrap(err, "Error processing hits")
	}

	id := res.ScrollId
	defer ClearScroll(cm.ClearScroll(), id)

	for {
		res, err = s.ScrollId(id).Do(context.Background())
		if err != nil {
			if err == io.EOF {
				break
			}

			return errors.Wrap(err, "Error querying elasticsearch when scrolling")
		}

		err = processHits(writer, res.Hits.Hits, config)
		if err != nil {
			return errors.Wrap(err, "Error processing hits")
		}
	}

	return nil
}

func processHits(writer io.Writer, searchHits []*elastic.SearchHit, config *WorkConfig) (err error) {
	for _, hit := range searchHits {
		v := make(map[string]interface{})

		err = json.Unmarshal(*hit.Source, &v)
		if err != nil {
			return errors.Wrap(err, "Error marshaling JSON to map[string]internface{}")
		}

		data, err := config.Rule.TemplateLog(v)
		if err != nil {
			return errors.Wrap(err, "Error calling body template")
		}

		_, err = writer.Write(data)
		if err != nil {
			return errors.Wrap(err, "Error writing templated hits to writer")
		}
	}

	return nil
}

func ClearScroll(service *elastic.ClearScrollService, id string) {
	if id != "" {
		_, err := service.ScrollId(id).Do(context.Background())
		if err != nil {
			util.Printf("%s\n", errors.Wrap(err, "Failed to clear scroll"))
		}
	}
}
