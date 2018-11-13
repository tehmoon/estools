package api

import (
	"../alert"
	"../storage"
	"sync"
	"../response"
	"github.com/tehmoon/errors"
	"net"
	"net/http"
	"time"
	"github.com/gorilla/mux"
)

type Manager struct {
	sync *sync.Mutex
	started bool
	config *ManagerConfig
	listener *net.TCPListener
	router *mux.Router
}

func NewManager(config *ManagerConfig) (manager *Manager, err error) {
	manager = &Manager{
		sync: &sync.Mutex{},
		config: config,
		router: mux.NewRouter(),
	}

	manager.router.
		HandleFunc("/alert/{id}/log", HTTPGetAlertIdLog(config.StorageManager)).
		Methods("GET")

	manager.router.
		HandleFunc("/response/tag/{tag}", HTTPGetResponseTag(config.ResponseManager)).
		Methods("GET")

	addr, err := net.ResolveTCPAddr("tcp", config.Listen)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to resolve listen address")
	}

	manager.listener, err = net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return nil, errors.Wrapf(err, "Error listening on address: %s", addr)
	}

	return manager, nil
}

type ManagerConfig struct {
	Listen string
	ResponseManager *response.Manager
	StorageManager *storage.Manager
	AlertManager *alert.Manager
}

func (m *Manager) Start() (error) {
	m.sync.Lock()

	if m.started {
		m.sync.Unlock()
		return errors.New("API Manager has already started")
	}

	m.started = true
	m.sync.Unlock()

	server := &http.Server{
		WriteTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		Handler: m.router,
	}

	server.Addr = m.listener.Addr().String()

	err := server.Serve(m.listener)
	if err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}
