package scheduler

import (
	"math/rand"
	"github.com/tehmoon/errors"
	"../rules"
	"../client"
	"../util"
	"../alert"
	"time"
	"sync"
)

type Manager struct {
	config *ManagerConfig
	cancel chan struct{}
	sync *sync.Mutex
	work chan *QueryConfig

	// Sets a lock during Run()
	scheduler []*RuleScheduler
}

// NOT THREAD SAFE
type RuleScheduler struct {
	Rule *rules.Rule
	lastScheduled *time.Time
	To time.Time
}

type ManagerConfig struct {
	RulesManager *rules.Manager
	ClientManager *client.Manager
	AlertManager *alert.Manager
	QueryDelay time.Duration
}

func NewManager(config *ManagerConfig) (manager *Manager) {
	manager = &Manager{
		config: config,
		sync: &sync.Mutex{},
		cancel: make(chan struct{}),
		work: make(chan *QueryConfig, 0),
		scheduler: make([]*RuleScheduler, 0),
	}

	go manager.start()

	return manager
}

func (m Manager) Run() {
	m.sync.Lock()
	defer m.sync.Unlock()

	now := time.Now()
	delayedNow := now.Add(-1 * m.config.QueryDelay)

	for _, rs := range m.scheduler {
		from := rs.Rule.From(&delayedNow)
		to := rs.Rule.To(&delayedNow)

		expired := rs.Expired(delayedNow, from, to)
		if ! expired {
			continue
		}

		q, err := rs.Rule.GenerateQuery(from, to)
		if err != nil {
			util.Println(errors.Wrapf(err, "Error creating query %q", rs.Rule.Name()).Error())
			continue
		}

		config := &QueryConfig{
			Rule: rs.Rule,
			ScheduledAt: now,
			From: from,
			To: to,
			ClientManager: m.config.ClientManager,
			Query: q,
			AlertManager: m.config.AlertManager,
		}

		util.Printf("Scheduling query %q [%q %q] time %q\n", rs.Rule.Name(), util.FormatTime(from), util.FormatTime(to), util.FormatTime(delayedNow))

		m.work <- config
	}
}

func (m Manager) Cancel() {
	m.cancel <- struct{}{}
}

func startWorker(work chan *QueryConfig, cancel chan struct{}) {
	for {
		select {
			case <- cancel:
				return
			case config := <- work:

				err := query(config)
				if err != nil {
					util.Println(errors.Wrapf(err, "Error running query %q", config.Rule.Name()).Error())
					continue
				}

				util.Printf("Done querying %q\n", config.Rule.Name())
		}
	}
}

func (m *Manager) start() {
	cancel := make(chan struct{}, 0)

	// TODO: make variable
	for i := 0; i < 10; i++ {
		go startWorker(m.work, cancel)
	}

	for _, rule := range m.config.RulesManager.Rules() {
		m.scheduler = append(m.scheduler, &RuleScheduler{
			Rule: rule,
		})
	}

	for {
		m.Run()

		ticker := time.NewTicker(time.Second)

		select {
			case <- m.cancel:
				// TODO: make variable
				for i := 0; i < 10; i++ {
					cancel <- struct{}{}
				}

				ticker.Stop()
				return
			case <- ticker.C:
				ticker.Stop()
		}
	}
}

// Return true if run_every is greater than now and if
// the from is greater than the last to. This avoid scheduling
// queries multiple times for the same range. Of course this is
// only the case when the to is not in the futur, otherwise, elasticsearch
// might not have the data.
// Also, it remembers when it has last called Expired()
// so it can do the math.
// NOT THREAD SAFE
func (rs *RuleScheduler) Expired(now, from, to time.Time) (bool) {
	// Initialize the first lastScheduled
	// I use a random number between 0 and max_wait_schedule
	// Then I subtract it to the now object actively "simulating"
	// a run in the past.
	// It is to spread the load so not too many queries run a the same time
	if rs.lastScheduled == nil {
		ra := rand.New(rand.NewSource(time.Now().UnixNano()))
		init := ra.Int63n(int64(rs.Rule.MaxWaitSchedule))

		now = now.Add(-1 * rs.Rule.RunEvery + time.Duration(init))
		rs.lastScheduled = &now
		rs.To = time.Unix(0, 0)
	}

	if now.Sub(*rs.lastScheduled) > rs.Rule.RunEvery {
		if rs.To.UnixNano() <= from.UnixNano() || to.UnixNano() > time.Now().UnixNano() {
			rs.To = to
			rs.lastScheduled = &now
			return true
		}
	}

	return false
}
