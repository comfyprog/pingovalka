package main

import (
	"log"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

func isOnline(addr string, size int, timeout time.Duration) bool {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		panic(err)
	}
	pinger.Count = 1
	pinger.Size = size
	pinger.Timeout = timeout
	pinger.Run()
	stats := pinger.Statistics()
	return stats.PacketsSent == stats.PacketsRecv
}

func pingHosts(hosts []Host, stopChan <-chan struct{}) chan Host {
	statusChan := make(chan Host, 5)
	for _, h := range hosts {
		go func(h Host, status chan<- Host, stop <-chan struct{}) {
			ticker := time.NewTicker(h.PingConfig.Interval)
			for {
				select {
				case <-ticker.C:
					var newStatus string
					if isOnline(h.Addr, h.PingConfig.Size, h.PingConfig.Timeout) {
						newStatus = online
					} else {
						newStatus = offline
					}
					if newStatus != h.Status {
						h.Status = newStatus
						status <- h
					}
				case <-stop:
					ticker.Stop()
					close(statusChan)
					return
				}
			}
		}(h, statusChan, stopChan)
	}
	return statusChan
}

type PingMux struct {
	mu       sync.Mutex
	hosts    []Host
	counter  int
	mainChan <-chan Host
	chans    map[int]chan Host
}

func NewPingMux(hosts []Host, mainChan <-chan Host) *PingMux {
	mux := PingMux{mainChan: mainChan, hosts: hosts}
	mux.chans = make(map[int]chan Host)

	return &mux
}

func (m *PingMux) AddSubscriber() (<-chan Host, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hostChan := make(chan Host, 1)
	counter := m.counter
	m.chans[counter] = hostChan
	delSubFunc := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		close(hostChan)
		delete(m.chans, counter)
	}

	m.counter++
	return hostChan, delSubFunc
}

func (m *PingMux) GetHosts() []Host {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hosts
}

func (m *PingMux) TransmitStatuses() {
	for host := range m.mainChan {
		log.Printf("%s [%s] changed status to '%s'", host.Name, host.Addr, host.Status)
		m.mu.Lock()
		for i := range m.hosts {
			if m.hosts[i].Id == host.Id {
				m.hosts[i].Status = host.Status
				break
			}
		}

		for _, c := range m.chans {
			c <- host
		}

		m.mu.Unlock()
	}
}
