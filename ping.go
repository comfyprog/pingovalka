package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func makePingStatusString(stat *probing.Statistics, t time.Duration) string {
	return fmt.Sprintf("ping %s at %s: %d packets transmitted, %d received, %.2f%% packet loss, time %v, avg rtt %v",
		stat.Addr, time.Now().Format(time.DateTime), stat.PacketsSent, stat.PacketsRecv, stat.PacketLoss, t.Round(10*time.Millisecond), stat.AvgRtt.Round(time.Millisecond))
}

func getStatus(addr string, count int, size int, timeout time.Duration, lastStatus string) (newStatus, newStatusText string) {
	pinger, err := probing.NewPinger(addr)
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

func pingHosts(hosts []Host, pingChan chan<- Host, stopChan <-chan struct{}, constantUpdates bool) {
	for _, h := range hosts {
		go func(h Host, status chan<- Host, stop <-chan struct{}) {
			ticker := time.NewTicker(h.PingConfig.Interval)
			for {
				select {
				case <-ticker.C:
					newStatus, newStatusText := getStatus(h.Addr, h.PingConfig.Count, h.PingConfig.Size, h.PingConfig.Timeout, h.Status)
					oldStatus := h.Status
					h.Status = newStatus
					h.StatusText = newStatusText
					if newStatus != oldStatus {
						h.StatusChangeTime = time.Now().Unix()
					}
					if constantUpdates || newStatus != oldStatus {
						status <- h
					}
				case <-stop:
					ticker.Stop()
					return
				}
			}
		}(h, pingChan, stopChan)
	}
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
