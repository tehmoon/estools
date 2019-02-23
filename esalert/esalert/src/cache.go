package main

import (
	"sync"
)

type CacheResponses struct {
	sync.RWMutex
	responses []*Response
}

func newCacheResponses() (*CacheResponses) {
	return &CacheResponses{
		responses: make([]*Response, 0),
	}
}

func (cr *CacheResponses) Update(responses []*Response) {
	cr.Lock()
	defer cr.Unlock()

	cr.update(responses)
}

func (cr *CacheResponses) update(responses []*Response) {
	tmp := make([]*Response, 0)

	for _, response := range responses {
		if cr.getId(response.Id) != nil {
			continue
		}

		tmp = append(tmp, response)
	}

	if len(tmp) == 0 {
		return
	}

	cr.responses = append(cr.responses, tmp...)
}

func (cr CacheResponses) GetId(id string) (*Response) {
	cr.RLock()
	defer cr.RUnlock()

	return cr.getId(id)
}

func (cr CacheResponses) getId(id string) (*Response) {
	for _, response := range cr.responses {
		if response.Id == id {
			return response
		}
	}

	return nil
}

func (cr *CacheResponses) Next(t NextResponse) (*Response) {
	cr.Lock()
	defer cr.Unlock()

	return cr.next(t)
}

func (cr *CacheResponses) next(t NextResponse) (*Response) {
	response := cr.getNextResponse(t)

	if response != nil {
		switch t {
			case NextResponseStart:
				response.handedOffStart = true
			case NextResponseStop:
				response.handedOffStop = true
		}
	}

	return response
}

func (cr *CacheResponses) getNextResponse(t NextResponse) (*Response) {
	for _, response := range cr.responses {
		switch t {
			case NextResponseStop:
				if ! response.handedOffStop && response.Expired() && response.execStart {
					return response
				}
			case NextResponseStart:
				if ! response.Expired() && ! response.handedOffStart {
					return response
				}
		}
	}

	return nil
}

type NextResponse int

const (
	NextResponseStart NextResponse = iota
	NextResponseStop
)
