package response

import (
	"encoding/json"
	"../rules"
	"time"
	"../util"
	"github.com/google/uuid"
	"sync"
	"reflect"
	"github.com/tehmoon/errors"
)

type Manager struct {
	tags map[string]Responses
	tagsCache map[string][]*json.RawMessage
	tagsSync *sync.RWMutex
}

type Responses []*Response

func (r Responses) exists(response *Response) (bool) {
	for _, rr := range r {
		if rr.Expired() && (reflect.DeepEqual(rr.Args, response.Args) && rr.Action() == response.Action()) {
			return true
		}
	}

	return false
}

type ManagerConfig struct {
}

type Response struct {
	GeneratedAt time.Time
	ExpireAt time.Time
	Id string
	config *ResponseConfig
	Args []string
}

type ResponsePayload struct {
	Id string `json:"id"`
	Args []string `json:"args"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpireAt time.Time `json:"expire_at"`
	Action string `json:"action"`
}

type ResponseConfig struct {
	TriggeredAt time.Time
	Rule *rules.Rule
	Count interface{}
	Value interface{}
}

type TemplateResponseRoot struct {
	Value interface{}
	Count interface{}
}

func NewManager(config *ManagerConfig) (manager *Manager, err error) {
	manager = &Manager{
		tagsSync: &sync.RWMutex{},
		tags: make(map[string]Responses, 0),
		tagsCache: make(map[string][]*json.RawMessage, 0),
	}

	return manager, nil
}

func (m Manager) Get(tag string) (responses []*Response, found bool) {
	responses, found = m.tags[tag]

	return
}

func (m Manager) GetCache(tag string) (payload []byte, found bool) {
	responses, found := m.tagsCache[tag]
	if ! found {
		return nil, false
	}

	payload, err := json.Marshal(responses)
	if err != nil {
		return nil, false
	}

	return payload, true
}

func (m *Manager) Add(config *ResponseConfig) (err error) {
	response, err := NewResponse(config)
	if err != nil {
		return err
	}

	m.tagsSync.Lock()
	defer m.tagsSync.Unlock()

	for _, tag := range response.Tags() {
		responses, found := m.tags[tag]
		if ! found {
			responses = make(Responses, 0)
		}

		cache, found := m.tagsCache[tag]
		if ! found {
			cache = make([]*json.RawMessage, 0)
		}

		exists := responses.exists(response)
		if exists {
			return errors.Errorf("Response already there and not expired for tag %q\n", tag)
		}

		p, err := response.GeneratePayload()
		if err != nil {
			err = errors.Wrapf(err, "Error generating payload for rule response %q", config.Rule.Name())
			return err
		}

		payload := json.RawMessage(p)

		responses = append(responses, response)
		cache = append(cache, &payload)

		m.tags[tag] = responses
		m.tagsCache[tag] = cache
	}

	return nil
}

func (r Response) Tags() (tags []string) {
	return r.config.Rule.Response.Tags
}

func (r Response) Action() (action string) {
	return r.config.Rule.Response.Action
}

func (r Response) Expired() (expired bool) {
	expireAt := r.config.TriggeredAt.Add(r.config.Rule.Response.Expire)

	if time.Now().Sub(expireAt) <= 0 {
		return true
	}

	return false
}

func (r Response) GeneratePayload() (payload []byte, err error) {
	rp := &ResponsePayload{
		Id: r.Id,
		Args: r.Args,
		Action: r.Action(),
		ExpireAt: r.ExpireAt,
		GeneratedAt: r.GeneratedAt,
	}

	return json.Marshal(rp)
}

func NewResponse(config *ResponseConfig) (response *Response, err error) {
	response = &Response{
		config: config,
		Args: make([]string, 0),
		GeneratedAt: time.Now(),
		Id: uuid.New().String(),
	}

	triggeredAt := response.config.TriggeredAt
	expire := response.config.Rule.Response.Expire

	response.ExpireAt = triggeredAt.Add(expire)

	for i, tmpl := range config.Rule.Response.Args {
		arg, err := util.TemplateToString(tmpl, &TemplateResponseRoot{
			Value: config.Value,
			Count: config.Count,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing argument number %d", i)
		}

		response.Args = append(response.Args, arg)
	}

	return response, nil
}
