package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

func makePingStatusString(stat *ping.Statistics, t time.Duration) string {
	return fmt.Sprintf("ping %s: %d packets transmitted, %d received, %.2f%% packet loss, time %v, avg rtt %v",
		stat.Addr, stat.PacketsSent, stat.PacketsRecv, stat.PacketLoss, t.Round(10*time.Millisecond), stat.AvgRtt.Round(10*time.Millisecond))
}

func getStatus(addr string, count int, size int, timeout time.Duration, lastStatus string) (newStatus, newStatusText string) {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		panic(err)
	}
	pinger.Count = count
	pinger.Size = size
	pinger.Timeout = timeout
	startTime := time.Now()
	err = pinger.Run()
	if err != nil {
		log.Printf("cannot ping %s: %s", addr, err)
		newStatus = lastStatus
		return
	}
	stats := pinger.Statistics()

	runDuration := time.Now().Sub(startTime)
	newStatusText = makePingStatusString(stats, runDuration)

	if stats.PacketsSent == 0 {
		newStatus = lastStatus
		return
	}

	if stats.PacketsRecv == 0 {
		newStatus = statusOffline
		return
	}

	if stats.PacketsRecv != stats.PacketsSent {
		newStatus = statusOnlineUnstable
		return
	}

	newStatus = statusOnline
	return
}

func pingHosts(hosts []Host, stopChan <-chan struct{}) chan Host {
	statusChan := make(chan Host, 5)
	for _, h := range hosts {
		go func(h Host, status chan<- Host, stop <-chan struct{}) {
			ticker := time.NewTicker(h.PingConfig.Interval)
			for {
				select {
				case <-ticker.C:
					newStatus, newStatusText := getStatus(h.Addr, h.PingConfig.Count, h.PingConfig.Size, h.PingConfig.Timeout, h.Status)
					if newStatus != h.Status {
						h.Status = newStatus
						h.StatusText = newStatusText
						h.StatusChangeTime = time.Now().Unix()
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
		delete(m.chans, counter)
	}

	m.counter++
	return hostChan, delSubFunc
}

func (m *PingMux) GetHosts() []Host {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := make([]Host, len(m.hosts))
	copy(s, m.hosts)
	return s
}

func (m *PingMux) TransmitStatuses() {
	for host := range m.mainChan {
		m.mu.Lock()
		log.Printf("%s [%s] changed status to '%s'", host.Name, host.Addr, host.Status)
		for i := range m.hosts {
			if m.hosts[i].Id == host.Id {
				m.hosts[i].Status = host.Status
				m.hosts[i].StatusText = host.StatusText
				m.hosts[i].StatusChangeTime = host.StatusChangeTime
				break
			}
		}

		for _, c := range m.chans {
			c <- host
		}

		m.mu.Unlock()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.chans {
		close(c)
	}
}
