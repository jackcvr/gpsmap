package main

import (
	"github.com/jackcvr/gpsmap/orm"
	"net/http"
	"sync"
	"time"
)

type PubSub struct {
	mu          sync.Mutex
	Subscribers map[http.ResponseWriter]chan orm.Record
}

func NewPubSub() *PubSub {
	return &PubSub{
		Subscribers: make(map[http.ResponseWriter]chan orm.Record),
	}
}

func (ps *PubSub) Subscribe(w http.ResponseWriter) chan orm.Record {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	sub := make(chan orm.Record, 1)
	ps.Subscribers[w] = sub
	return sub
}

func (ps *PubSub) Unsubscribe(w http.ResponseWriter) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	sub := ps.Subscribers[w]
	close(sub)
	delete(ps.Subscribers, w)
}

func (ps *PubSub) Publish(r orm.Record) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, sub := range ps.Subscribers {
		sub <- r
	}
}

func RunPeriodic(d time.Duration, f func()) {
	t := time.NewTicker(d)
	for {
		<-t.C
		f()
	}
}
